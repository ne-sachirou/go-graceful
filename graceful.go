// Package graceful provides utilities for shutting down servers gracefully.
package graceful

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"time"
)

// Config represents configuration parameters for the Server.
type Config struct {
	Signals         []os.Signal
	ShutdownTimeout time.Duration
}

// setDefault resets c with default parameters.
func (c Config) setDefault() Config {
	if len(c.Signals) == 0 {
		c.Signals = []os.Signal{os.Interrupt}
	}
	return c
}

// Server is the interface that wraps the Serve and Shutdown methods.
type Server interface {
	Serve(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// Servers is the collection of the servers will be shutting down at the same time.
type Servers struct {
	Servers []Server
}

// Graceful runs all servers contained in s, then waits signals.
// When receive an expected signal, s stops all servers gracefully.
func (s Servers) Graceful(ctx context.Context, cfg Config) error {
	cfg = cfg.setDefault()

	ctx, stop := signal.NotifyContext(ctx, cfg.Signals...)
	defer stop()

	ctx, cancel := context.WithCancelCause(ctx)

	for _, srv := range s.Servers {
		go func(ctx context.Context, srv Server) {
			if err := srv.Serve(ctx); err != nil {
				cancel(err)
			}
		}(ctx, srv)
	}

	// 終了処理

	<-ctx.Done()
	if err := context.Cause(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return errors.Join(errors.New("failed to start servers"), err)
	}

	ctx, cancelT := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelT()

	var wg sync.WaitGroup

	var shutdownErr error = nil
	var shutdownErrMu sync.Mutex

	for _, srv := range s.Servers {
		wg.Add(1)
		go func(ctx context.Context, srv Server) {
			defer wg.Done()
			if err := srv.Shutdown(ctx); err != nil {
				shutdownErrMu.Lock()
				defer shutdownErrMu.Unlock()
				shutdownErr = errors.Join(shutdownErr, err)
			}
		}(ctx, srv)
	}

	wg.Wait()
	if err := context.Cause(ctx); shutdownErr != nil || (err != nil && !errors.Is(err, context.Canceled)) {
		return errors.Join(errors.New("failed to shutdown gracefully"), shutdownErr, err)
	}
	return nil
}
