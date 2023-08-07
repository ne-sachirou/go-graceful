package main

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/ne-sachirou/go-graceful/graceful"
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
	slog.InfoCtx(ctx, "start")
	if err := (graceful.Servers{
		Logger:  slog.Default(),
		Servers: []graceful.Server{&ExampleServer{}},
	}.Graceful(
		ctx,
		graceful.GracefulConfig{ShutdownTimeout: time.Second},
	)); err != nil {
		slog.ErrorCtx(ctx, err.Error())
		os.Exit(1)
	}
	slog.InfoCtx(ctx, "finish")
	os.Exit(0)
}
