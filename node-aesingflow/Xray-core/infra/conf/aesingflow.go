package conf

import (
	"strings"

	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/proxy/aesingflow"
	"google.golang.org/protobuf/proto"
)

// AesingFlowTLSConfig configures the verified TLS client used for AesingFlow.
// A private CA file must contain public PEM certificates only.
type AesingFlowTLSConfig struct {
	ServerName string `json:"serverName"`
	CAFile     string `json:"caFile"`
}

// AesingFlowClientConfig is the JSON configuration for protocol "aesingflow".
// It transports TCP only and intentionally has no insecure TLS option.
type AesingFlowClientConfig struct {
	Server                        string               `json:"server"`
	ServerPort                    uint32               `json:"serverPort"`
	Token                         string               `json:"token"`
	TLS                           *AesingFlowTLSConfig `json:"tls"`
	MaxStreams                    uint32               `json:"maxStreams"`
	BrutalBps                     uint64               `json:"brutalBps"`
	DisableBrutal                 bool                 `json:"disableBrutal"`
	BrutalDisableLossCompensation bool                 `json:"brutalDisableLossCompensation"`
}

// AesingFlowUserConfig defines one Remnawave-managed inbound user. Certificate
// material deliberately does not live here: it belongs to tlsSettings.
type AesingFlowUserConfig struct {
	ID    string `json:"id"`
	Token string `json:"token"`
	Email string `json:"email"`
	Level uint32 `json:"level"`
}

// AesingFlowServerConfig is the JSON configuration for an AesingFlow inbound.
// QUIC/TLS is selected with streamSettings.network = "aesingflow" and the
// standard streamSettings.tlsSettings object.
type AesingFlowServerConfig struct {
	Clients                       []*AesingFlowUserConfig `json:"clients"`
	MaxStreams                    uint32                  `json:"maxStreams"`
	BrutalBps                     uint64                  `json:"brutalBps"`
	DisableBrutal                 bool                    `json:"disableBrutal"`
	BrutalDisableLossCompensation bool                    `json:"brutalDisableLossCompensation"`
}

// Build implements Buildable.
func (c *AesingFlowServerConfig) Build() (proto.Message, error) {
	if len(c.Clients) == 0 {
		return nil, errors.New("AesingFlow inbound requires at least one client")
	}

	config := &aesingflow.ServerConfig{
		MaxStreams:                    c.MaxStreams,
		BrutalBps:                     c.BrutalBps,
		DisableBrutal:                 c.DisableBrutal,
		BrutalDisableLossCompensation: c.BrutalDisableLossCompensation,
		Clients:                       make([]*aesingflow.User, 0, len(c.Clients)),
	}
	seenTokens := make(map[string]struct{}, len(c.Clients))
	for _, client := range c.Clients {
		if client == nil || strings.TrimSpace(client.ID) == "" {
			return nil, errors.New("AesingFlow inbound client id is required")
		}
		if strings.TrimSpace(client.Token) == "" {
			return nil, errors.New("AesingFlow inbound client token is required")
		}
		if _, duplicate := seenTokens[client.Token]; duplicate {
			return nil, errors.New("AesingFlow inbound client tokens must be unique")
		}
		seenTokens[client.Token] = struct{}{}
		email := client.Email
		if strings.TrimSpace(email) == "" {
			email = client.ID
		}
		config.Clients = append(config.Clients, &aesingflow.User{
			Id: client.ID, Token: client.Token, Email: email, Level: client.Level,
		})
	}
	return config, nil
}

// Build implements Buildable.
func (c *AesingFlowClientConfig) Build() (proto.Message, error) {
	if strings.TrimSpace(c.Server) == "" {
		return nil, errors.New("AesingFlow server is required")
	}
	if c.ServerPort == 0 || c.ServerPort > 65535 {
		return nil, errors.New("AesingFlow serverPort must be between 1 and 65535")
	}
	if strings.TrimSpace(c.Token) == "" {
		return nil, errors.New("AesingFlow token is required")
	}

	config := &aesingflow.Config{
		Server:                        c.Server,
		ServerPort:                    c.ServerPort,
		Token:                         c.Token,
		MaxStreams:                    c.MaxStreams,
		BrutalBps:                     c.BrutalBps,
		DisableBrutal:                 c.DisableBrutal,
		BrutalDisableLossCompensation: c.BrutalDisableLossCompensation,
	}
	if c.TLS != nil {
		config.ServerName = c.TLS.ServerName
		config.CaFile = c.TLS.CAFile
	}
	return config, nil
}
