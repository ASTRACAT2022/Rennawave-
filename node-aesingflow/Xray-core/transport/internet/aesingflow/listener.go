package aesingflow

import (
	"context"
	stdtls "crypto/tls"
	stdnet "net"
	"strconv"
	"sync"
	"time"

	flow "github.com/ASTRACAT2022/aesingflow/pkg/aesingflow"
	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/errors"
	xnet "github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/transport/internet"
	"github.com/xtls/xray-core/transport/internet/stat"
	xraytls "github.com/xtls/xray-core/transport/internet/tls"
)

type listener struct {
	server flow.Server
	cancel context.CancelFunc
	once   sync.Once
}

func (l *listener) Addr() stdnet.Addr { return l.server.Addr() }

func (l *listener) Close() error {
	l.once.Do(l.cancel)
	return l.server.Close()
}

// Listen starts one AesingFlow QUIC listener. It uses Xray's TLS loader once,
// so there is no secondary certificate store or TLS termination layer.
func Listen(ctx context.Context, address xnet.Address, port xnet.Port, settings *internet.MemoryStreamConfig, addConn internet.ConnHandler) (internet.Listener, error) {
	if address.Family().IsDomain() {
		return nil, errors.New("AesingFlow listen address must be an IP address")
	}
	config, ok := inboundConfigFromContext(ctx)
	if !ok || config.Authenticator == nil {
		return nil, errors.New("AesingFlow inbound authentication is not configured")
	}
	xrayConfig := xraytls.ConfigFromStreamSettings(settings)
	if err := validateTLS(xrayConfig); err != nil {
		return nil, errors.New("invalid AesingFlow TLS configuration").Base(err)
	}
	tlsConfig := xrayConfig.GetTLSConfig()
	// Xray's standard TLS listener normally serves certificates through
	// GetCertificate. The AesingFlow QUIC library validates the server config
	// using tls.Config.Certificates directly, so expose the same certificates
	// there as well. This still uses the standard streamSettings.tlsSettings
	// loader and does not introduce a second certificate store.
	certificates := xrayConfig.BuildCertificates()
	if len(certificates) == 0 {
		return nil, errors.New("AesingFlow TLS configuration contains no usable certificate")
	}
	tlsConfig.Certificates = make([]stdtls.Certificate, 0, len(certificates))
	for _, certificate := range certificates {
		if certificate != nil {
			tlsConfig.Certificates = append(tlsConfig.Certificates, *certificate)
		}
	}
	if len(tlsConfig.Certificates) == 0 {
		return nil, errors.New("AesingFlow TLS configuration contains no usable certificate")
	}
	// validateTLS has already required TLS 1.3 and the expected ALPN. The
	// AesingFlow library uses this exact standard TLS config for QUIC.
	server, err := flow.NewServer(flow.ServerConfig{
		Address:                       stdnet.JoinHostPort(address.String(), strconv.Itoa(int(port))),
		TLSConfig:                     tlsConfig,
		Authenticator:                 config.Authenticator,
		MaxStreamsPerClient:           config.MaxStreams,
		BrutalSendRate:                config.BrutalBps,
		DisableBrutal:                 config.DisableBrutal,
		BrutalDisableLossCompensation: config.BrutalDisableLossCompensation,
	})
	if err != nil {
		return nil, errors.New("failed to start AesingFlow QUIC listener").Base(err)
	}
	listenCtx, cancel := context.WithCancel(ctx)
	l := &listener{server: server, cancel: cancel}
	go l.acceptConnections(listenCtx, addConn)
	return l, nil
}

func (l *listener) acceptConnections(ctx context.Context, addConn internet.ConnHandler) {
	for {
		conn, err := l.server.Accept(ctx)
		if err != nil {
			if ctx.Err() == nil {
				errors.LogInfoInner(ctx, err, "AesingFlow accept failed")
			}
			return
		}
		go l.acceptStreams(ctx, conn, addConn)
	}
}

func (l *listener) acceptStreams(ctx context.Context, conn flow.Connection, addConn internet.ConnHandler) {
	for {
		stream, err := conn.AcceptStream(ctx)
		if err != nil {
			return
		}
		addConn(&streamConn{stream: stream, local: flowAddr("aesingflow"), remote: flowAddr("aesingflow-client")})
	}
}

type streamConn struct {
	stream flow.StreamSession
	local  stdnet.Addr
	remote stdnet.Addr
}

func (c *streamConn) Read(p []byte) (int, error)         { return c.stream.Read(p) }
func (c *streamConn) Write(p []byte) (int, error)        { return c.stream.Write(p) }
func (c *streamConn) Close() error                       { return c.stream.Close() }
func (c *streamConn) LocalAddr() stdnet.Addr             { return c.local }
func (c *streamConn) RemoteAddr() stdnet.Addr            { return c.remote }
func (c *streamConn) SetDeadline(t time.Time) error      { return c.stream.SetDeadline(t) }
func (c *streamConn) SetReadDeadline(t time.Time) error  { return c.stream.SetReadDeadline(t) }
func (c *streamConn) SetWriteDeadline(t time.Time) error { return c.stream.SetWriteDeadline(t) }

var _ stat.Connection = (*streamConn)(nil)

type flowAddr string

func (a flowAddr) Network() string { return protocolName }
func (a flowAddr) String() string  { return string(a) }

func init() {
	common.Must(internet.RegisterTransportListener(protocolName, Listen))
}
