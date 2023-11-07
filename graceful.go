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

// Server is the interface that wraps the Serve and Shutdown methods.
type Server interface {
	Serve(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// Servers is the collection of the servers will be shutting down at the same time.
type Servers struct {
	Servers []Server
}

// GracefulOpts represents configuration parameters for the Server.
// In default, GracefulOpts.Signals is []os.Signal{os.Interrupt} and GracefulOpts.ShutdownTimeout is 0.
type GracefulOpts struct {
	Signals         []os.Signal
	ShutdownTimeout time.Duration
}

func defaultGracefulOpts() GracefulOpts {
	return GracefulOpts{
		Signals:         []os.Signal{os.Interrupt},
		ShutdownTimeout: 0,
	}
}

// GracefulSignals sets signals to be received.
func GracefulSignals(signals ...os.Signal) func(*GracefulOpts) {
	return func(o *GracefulOpts) { o.Signals = signals }
}

// GracefulShutdownTimeout sets timeout for shutdown.
func GracefulShutdownTimeout(timeout time.Duration) func(*GracefulOpts) {
	return func(o *GracefulOpts) { o.ShutdownTimeout = timeout }
}

// Graceful runs all servers contained in s, then waits signals.
// When receive an expected signal, s stops all servers gracefully.
func (s Servers) Graceful(ctx context.Context, options ...func(*GracefulOpts)) error {
	opts := defaultGracefulOpts()
	for _, f := range options {
		f(&opts)
	}

	ctx, stop := signal.NotifyContext(ctx, opts.Signals...)
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

	ctx, cancelT := context.WithTimeout(context.Background(), opts.ShutdownTimeout)
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
