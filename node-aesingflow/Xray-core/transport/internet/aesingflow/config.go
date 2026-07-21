// Package aesingflow registers AesingFlow's QUIC listener as an Xray stream
// transport. Its TLS configuration comes from the standard TLS stream config.
package aesingflow

import (
	"context"

	flow "github.com/ASTRACAT2022/aesingflow/pkg/aesingflow"
	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/transport/internet"
)

const protocolName = "aesingflow"

// Config is intentionally empty: certificates, private keys and TLS options
// are all represented by standard streamSettings.tlsSettings.
type Config struct{}

// InboundConfig is passed by the inbound protocol to the transport at startup.
// It contains authentication and flow-control state only.
type InboundConfig struct {
	Authenticator                 flow.Authenticator
	MaxStreams                    int
	BrutalBps                     uint64
	DisableBrutal                 bool
	BrutalDisableLossCompensation bool
}

type inboundConfigKey struct{}

func ContextWithInboundConfig(ctx context.Context, config InboundConfig) context.Context {
	return context.WithValue(ctx, inboundConfigKey{}, config)
}

func inboundConfigFromContext(ctx context.Context) (InboundConfig, bool) {
	config, ok := ctx.Value(inboundConfigKey{}).(InboundConfig)
	return config, ok
}

func init() {
	common.Must(internet.RegisterProtocolConfigCreator(protocolName, func() interface{} { return &Config{} }))
}
