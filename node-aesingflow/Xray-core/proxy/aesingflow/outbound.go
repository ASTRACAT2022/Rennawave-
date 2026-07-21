// Package aesingflow implements AesingFlow as a TCP-only Xray outbound.
package aesingflow

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	stdnet "net"
	"os"
	"strconv"
	"strings"

	flow "github.com/ASTRACAT2022/aesingflow/pkg/aesingflow"
	flowproxy "github.com/ASTRACAT2022/aesingflow/proxy"
	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/buf"
	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/session"
	"github.com/xtls/xray-core/common/signal"
	"github.com/xtls/xray-core/common/task"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/policy"
	"github.com/xtls/xray-core/features/routing"
	"github.com/xtls/xray-core/transport"
	"github.com/xtls/xray-core/transport/internet"
	aftransport "github.com/xtls/xray-core/transport/internet/aesingflow"
	"github.com/xtls/xray-core/transport/internet/stat"
)

// Handler owns one reusable AesingFlow dialer for an outbound profile.
// The dialer multiplexes all TCP streams over one QUIC connection.
type Handler struct {
	dialer        *flowproxy.Dialer
	policyManager policy.Manager
}

// Server accepts authenticated AesingFlow streams supplied by the dedicated
// QUIC transport. TLS configuration is intentionally not part of this proxy.
type Server struct {
	config        *ServerConfig
	inboundConfig aftransport.InboundConfig
}

// NewServer builds the token authenticator for an AesingFlow inbound.
// Remnawave starts managed inbounds with an empty clients array and then
// synchronizes users later, so an empty token set must still create a valid
// listener. Until users are added, the authenticator rejects every client.
func NewServer(_ context.Context, config *ServerConfig) (*Server, error) {
	tokens := make([]flow.Token, 0, len(config.Clients))
	for _, client := range config.Clients {
		if client == nil || strings.TrimSpace(client.Token) == "" {
			return nil, errors.New("AesingFlow inbound client token is required")
		}
		subject := client.Email
		if subject == "" {
			subject = client.Id
		}
		tokens = append(tokens, flow.Token{Value: client.Token, Subject: subject})
	}
	return &Server{config: config, inboundConfig: aftransport.InboundConfig{
		Authenticator:                 &flow.StaticAuthenticator{Tokens: tokens},
		MaxStreams:                    int(config.MaxStreams),
		BrutalBps:                     config.BrutalBps,
		DisableBrutal:                 config.DisableBrutal,
		BrutalDisableLossCompensation: config.BrutalDisableLossCompensation,
	}}, nil
}

// AesingFlowInboundConfig supplies only authentication and flow limits to the
// transport. Certificate paths remain exclusively in tlsSettings.
func (s *Server) AesingFlowInboundConfig() aftransport.InboundConfig { return s.inboundConfig }

func (s *Server) Network() []net.Network { return []net.Network{net.Network_TCP} }

// Process dispatches the authenticated AesingFlow CONNECT stream through the
// normal Xray routing pipeline.
func (s *Server) Process(ctx context.Context, network net.Network, conn stat.Connection, dispatcher routing.Dispatcher) error {
	if network != net.Network_TCP {
		return errors.New("AesingFlow inbound supports TCP streams only")
	}
	inbound := session.InboundFromContext(ctx)
	inbound.Name = "aesingflow"
	target, err := flowproxy.ReadRequest(conn)
	if err != nil {
		return errors.New("invalid AesingFlow CONNECT request").Base(err)
	}
	dest, err := net.ParseDestination("tcp:" + target.Address())
	if err != nil {
		_ = flowproxy.WriteResponse(conn, false)
		return errors.New("invalid AesingFlow destination").Base(err)
	}
	if err := flowproxy.WriteResponse(conn, true); err != nil {
		return errors.New("failed to write AesingFlow CONNECT response").Base(err)
	}
	return dispatcher.DispatchLink(ctx, dest, &transport.Link{Reader: buf.NewReader(conn), Writer: buf.NewWriter(conn)})
}

// New creates an AesingFlow outbound with certificate verification enabled.
func New(ctx context.Context, config *Config) (*Handler, error) {
	if strings.TrimSpace(config.Server) == "" {
		return nil, errors.New("AesingFlow server is required")
	}
	if config.ServerPort == 0 || config.ServerPort > 65535 {
		return nil, errors.New("AesingFlow server port must be between 1 and 65535")
	}
	if strings.TrimSpace(config.Token) == "" {
		return nil, errors.New("AesingFlow token is required")
	}

	serverName := config.ServerName
	if serverName == "" && net.ParseAddress(config.Server).Family().IsDomain() {
		serverName = config.Server
	}
	if serverName == "" {
		return nil, errors.New("AesingFlow tls.serverName is required when server is an IP address")
	}

	tlsConfig, err := newTLSConfig(serverName, config.CaFile)
	if err != nil {
		return nil, err
	}
	client, err := flow.NewClient(flow.ClientConfig{
		Address:                       stdnet.JoinHostPort(config.Server, strconv.Itoa(int(config.ServerPort))),
		TLSConfig:                     tlsConfig,
		Token:                         config.Token,
		MaxStreams:                    int(config.MaxStreams),
		BrutalSendRate:                config.BrutalBps,
		DisableBrutal:                 config.DisableBrutal,
		BrutalDisableLossCompensation: config.BrutalDisableLossCompensation,
	})
	if err != nil {
		return nil, errors.New("failed to create AesingFlow client").Base(err)
	}
	dialer, err := flowproxy.NewDialer(flowproxy.DialerConfig{Client: client})
	if err != nil {
		return nil, errors.New("failed to create AesingFlow dialer").Base(err)
	}

	v := core.MustFromContext(ctx)
	return &Handler{
		dialer:        dialer,
		policyManager: v.GetFeature(policy.ManagerType()).(policy.Manager),
	}, nil
}

func newTLSConfig(serverName, caFile string) (*tls.Config, error) {
	config := &tls.Config{
		MinVersion: tls.VersionTLS13,
		ServerName: serverName,
	}
	if caFile == "" {
		return config, nil
	}

	pem, err := os.ReadFile(caFile)
	if err != nil {
		return nil, errors.New("failed to read AesingFlow CA file").Base(err)
	}
	pool, err := x509.SystemCertPool()
	if err != nil {
		pool = x509.NewCertPool()
	}
	if pool == nil {
		pool = x509.NewCertPool()
	}
	if !pool.AppendCertsFromPEM(pem) {
		return nil, errors.New("AesingFlow CA file contains no PEM certificates")
	}
	config.RootCAs = pool
	return config, nil
}

// Process implements proxy.Outbound. UDP is deliberately rejected because the
// AesingFlow integration currently transports TCP CONNECT streams only.
func (h *Handler) Process(ctx context.Context, link *transport.Link, _ internet.Dialer) error {
	outbounds := session.OutboundsFromContext(ctx)
	ob := outbounds[len(outbounds)-1]
	if !ob.Target.IsValid() {
		return errors.New("target not specified")
	}
	if ob.Target.Network != net.Network_TCP {
		return errors.New("AesingFlow outbound supports TCP only; UDP is not supported")
	}
	ob.Name = "aesingflow"

	conn, err := h.dialer.DialContext(ctx, "tcp", ob.Target.NetAddr())
	if err != nil {
		return errors.New("failed to open AesingFlow stream").Base(err)
	}
	defer conn.Close()

	p := h.policyManager.ForLevel(0)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	timer := signal.CancelAfterInactivity(ctx, cancel, p.Timeouts.ConnectionIdle)

	requestDone := func() error {
		defer timer.SetTimeout(p.Timeouts.DownlinkOnly)
		return buf.Copy(link.Reader, buf.NewWriter(conn), buf.UpdateActivity(timer))
	}
	responseDone := func() error {
		defer timer.SetTimeout(p.Timeouts.UplinkOnly)
		return buf.Copy(buf.NewReader(conn), link.Writer, buf.UpdateActivity(timer))
	}
	if err := task.Run(ctx, requestDone, task.OnSuccess(responseDone, task.Close(link.Writer))); err != nil {
		return errors.New("AesingFlow connection ends").Base(err)
	}
	return nil
}

// Close releases the shared QUIC connection during outbound stop or reload.
func (h *Handler) Close() error {
	return h.dialer.Close()
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*Config))
	}))
	common.Must(common.RegisterConfig((*ServerConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewServer(ctx, config.(*ServerConfig))
	}))
}
