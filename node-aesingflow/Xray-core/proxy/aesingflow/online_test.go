package aesingflow

import (
	stdnet "net"
	"testing"

	"github.com/xtls/xray-core/transport/internet/stat"
)

type subjectTestConnection struct {
	stdnet.Conn
	subject string
}

func (c *subjectTestConnection) AesingFlowSubject() string { return c.subject }

func TestAuthenticatedSubjectUnwrapsStatsConnection(t *testing.T) {
	client, server := stdnet.Pipe()
	defer client.Close()
	defer server.Close()

	conn := &subjectTestConnection{Conn: client, subject: "user-123"}
	wrapped := &stat.CounterConnection{Connection: conn}
	if got := authenticatedSubject(wrapped); got != "user-123" {
		t.Fatalf("authenticatedSubject() = %q, want user-123", got)
	}
}
