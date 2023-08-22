package main

import (
	"context"
	_ "errors"
	"os"
	"time"

	"github.com/ne-sachirou/go-graceful"
	"golang.org/x/exp/slog"
)

type ExampleServer struct{}

func (s *ExampleServer) Serve(ctx context.Context) error {
	slog.InfoCtx(ctx, "serve")
	return nil
	//return errors.New("can't serve")
}

func (s *ExampleServer) Shutdown(ctx context.Context) error {
	slog.InfoCtx(ctx, "shutdown")
	return nil
	//return errors.New("can't shutdown")
}

func main() {
	ctx := context.Background()
	srv := graceful.Servers{
		Servers: []graceful.Server{&ExampleServer{}},
	}
	cfg := graceful.Config{ShutdownTimeout: time.Second}
	if err := srv.Graceful(ctx, cfg); err != nil {
		slog.ErrorCtx(ctx, err.Error())
		os.Exit(1)
	}
}
