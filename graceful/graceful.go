package graceful

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"time"

	"golang.org/x/exp/slog"
)

type Server interface {
	Serve(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

type Servers struct {
	Logger  *slog.Logger
	Servers []Server
}

type GracefulConfig struct {
	ShutdownTimeout time.Duration
}

// Servers_Graceful
func (s Servers) Graceful(ctx context.Context, cfg GracefulConfig) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
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
		//s.Logger.Error("failed to start servers", slog.String("err", err.Error()))
		return errors.Join(errors.New("failed to start servers"), err)
	}

	ctx, cancelT := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelT()

	var wg sync.WaitGroup

	for _, srv := range s.Servers {
		wg.Add(1)
		go func(ctx context.Context, srv Server) {
			defer wg.Done()
			if err := srv.Shutdown(ctx); err != nil {
				s.Logger.ErrorCtx(ctx, "failed to shutdown the server", slog.String("err", err.Error()))
			}
		}(ctx, srv)
	}

	wg.Wait()
	if err := context.Cause(ctx); err != nil && !errors.Is(err, context.Canceled) {
		//s.Logger.ErrorCtx(ctx, "failed to shutdown gracefully", slog.String("err", err.Error()))
		return errors.Join(errors.New("failed to shutdown gracefully"), err)
	}
	return nil
}
