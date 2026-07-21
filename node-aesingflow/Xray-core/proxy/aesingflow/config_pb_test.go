package aesingflow

import (
	"testing"

	"google.golang.org/protobuf/proto"
)

func TestServerConfigRoundTrip(t *testing.T) {
	want := &ServerConfig{
		Clients:    []*User{{Id: "user", Token: "secret", Email: "user", Level: 1}},
		MaxStreams: 256,
		BrutalBps:  250000000,
	}
	encoded, err := proto.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	got := new(ServerConfig)
	if err := proto.Unmarshal(encoded, got); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(want, got) {
		t.Fatalf("server config did not round-trip: got %#v", got)
	}
}
