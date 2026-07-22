package aesingflow

import (
	stdnet "net"
	"testing"
)

func TestTCPAddrConvertsQUICUDPAddress(t *testing.T) {
	input := &stdnet.UDPAddr{IP: stdnet.ParseIP("203.0.113.42"), Port: 4433}
	got := tcpAddr(input)
	if !got.IP.Equal(input.IP) || got.Port != input.Port {
		t.Fatalf("tcpAddr() = %v, want %v:%d", got, input.IP, input.Port)
	}
}

func TestTCPAddrAlwaysReturnsTCPAddress(t *testing.T) {
	got := tcpAddr(flowAddrForTest("aesingflow"))
	if got == nil {
		t.Fatal("tcpAddr() returned nil")
	}
}

type flowAddrForTest string

func (a flowAddrForTest) Network() string { return "aesingflow" }
func (a flowAddrForTest) String() string  { return string(a) }
