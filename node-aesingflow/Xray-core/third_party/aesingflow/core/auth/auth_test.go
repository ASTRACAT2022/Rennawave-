package auth

import (
	"context"
	aferrors "github.com/ASTRACAT2022/aesingflow/core/errors"
	"sync"
	"testing"
	"time"
)

func TestStaticAuthenticator(t *testing.T) {
	a := &StaticAuthenticator{Tokens: []Token{{Value: "secret", Subject: "test"}}}
	var n [16]byte
	n[0] = 1
	r := Request{"secret", n, time.Now()}
	if _, e := a.Authenticate(context.Background(), r); e != nil {
		t.Fatal(e)
	}
	if aferrors.CodeOf(mustErr(a, r)) != aferrors.ReplayDetected {
		t.Fatal("expected replay")
	}
}
func mustErr(a *StaticAuthenticator, r Request) error {
	_, e := a.Authenticate(context.Background(), r)
	return e
}
func TestExpired(t *testing.T) {
	a := &StaticAuthenticator{Tokens: []Token{{Value: "s", ExpiresAt: time.Now().Add(-time.Second)}}}
	_, e := a.Authenticate(context.Background(), Request{"s", [16]byte{2}, time.Now()})
	if aferrors.CodeOf(e) != aferrors.AuthExpired {
		t.Fatal(e)
	}
}

func TestStaticAuthenticatorConcurrent(t *testing.T) {
	a := &StaticAuthenticator{Tokens: []Token{{Value: "secret", Subject: "test"}}}
	const requests = 128
	errs := make(chan error, requests)
	var wg sync.WaitGroup
	for i := range requests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var nonce [16]byte
			nonce[0] = byte(i)
			nonce[1] = byte(i >> 8)
			_, err := a.Authenticate(context.Background(), Request{Token: "secret", Nonce: nonce, Timestamp: time.Now()})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}
