package proxy

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"runtime"
	"sync"
	"time"

	aferrors "github.com/ASTRACAT2022/aesingflow/core/errors"
	"github.com/ASTRACAT2022/aesingflow/pkg/aesingflow"
)

// ServerConfig configures the exit side of an AesingFlow TCP proxy.
type ServerConfig struct {
	DialTimeout          time.Duration
	RequestTimeout       time.Duration
	MaxConcurrentStreams int
	AcceptWorkers        int
	Logger               *slog.Logger
}

// Serve accepts AesingFlow connections and proxies each stream to its requested
// TCP endpoint. Access to this service must be protected with a strong token.
func Serve(ctx context.Context, listener aesingflow.Server, cfg ServerConfig) error {
	if listener == nil {
		return fmt.Errorf("proxy: AesingFlow server is required")
	}
	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = 10 * time.Second
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 10 * time.Second
	}
	if cfg.MaxConcurrentStreams <= 0 {
		cfg.MaxConcurrentStreams = 4096
	}
	if cfg.AcceptWorkers <= 0 {
		cfg.AcceptWorkers = runtime.GOMAXPROCS(0)
		// quic-go's completed-handshake queue is deliberately small. Keep enough
		// application-handshake workers to drain a connection burst before that
		// queue starts refusing otherwise valid clients.
		if cfg.AcceptWorkers < 64 {
			cfg.AcceptWorkers = 64
		}
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	connections := make(chan aesingflow.Connection, cfg.AcceptWorkers)
	var workers sync.WaitGroup
	for range cfg.AcceptWorkers {
		workers.Add(1)
		go func() {
			defer workers.Done()
			acceptConnections(ctx, listener, cfg, connections)
		}()
	}
	workersDone := make(chan struct{})
	go func() {
		workers.Wait()
		close(workersDone)
	}()
	defer func() {
		if ctx.Err() == nil {
			_ = listener.Close()
		}
		<-workersDone
	}()

	streamSlots := make(chan struct{}, cfg.MaxConcurrentStreams)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-workersDone:
			return nil
		case conn := <-connections:
			go serveConnection(ctx, conn, cfg, streamSlots)
		}
	}
}

func acceptConnections(ctx context.Context, listener aesingflow.Server, cfg ServerConfig, connections chan<- aesingflow.Connection) {
	for {
		conn, err := listener.Accept(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			var protocolErr *aferrors.Error
			if errors.As(err, &protocolErr) {
				if protocolErr.Code == aferrors.ShuttingDown {
					return
				}
				cfg.Logger.Debug("AesingFlow connection rejected", "code", protocolErr.Code, "error", err)
				continue
			}
			cfg.Logger.Warn("AesingFlow accept failed; retrying", "error", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(25 * time.Millisecond):
			}
			continue
		}
		select {
		case connections <- conn:
		case <-ctx.Done():
			_ = conn.CloseWithError(0, "proxy stopped")
			return
		}
	}
}

func serveConnection(ctx context.Context, conn aesingflow.Connection, cfg ServerConfig, streamSlots chan struct{}) {
	defer conn.CloseWithError(0, "proxy client disconnected")
	for {
		stream, err := conn.AcceptStream(ctx)
		if err != nil {
			return
		}
		select {
		case streamSlots <- struct{}{}:
			go func() {
				defer func() { <-streamSlots }()
				serveStream(ctx, stream, cfg)
			}()
		case <-ctx.Done():
			_ = stream.Close()
			return
		}
	}
}

func serveStream(ctx context.Context, stream aesingflow.StreamSession, cfg ServerConfig) {
	defer stream.Close()
	_ = stream.SetDeadline(time.Now().Add(cfg.RequestTimeout))
	target, err := readRequest(stream)
	if err != nil {
		cfg.Logger.Debug("invalid proxy request", "error", err)
		return
	}
	_ = stream.SetDeadline(time.Time{})
	dialCtx, cancel := context.WithTimeout(ctx, cfg.DialTimeout)
	remote, err := (&net.Dialer{}).DialContext(dialCtx, "tcp", target.Address())
	cancel()
	if err != nil {
		_ = writeResponse(stream, statusFailure)
		cfg.Logger.Debug("proxy target connection failed", "target", target.Address(), "error", err)
		return
	}
	defer remote.Close()
	if err = writeResponse(stream, statusOK); err != nil {
		return
	}
	cfg.Logger.Debug("proxy target connection opened", "target", target.Address())
	copyBoth(remote, stream)
}
