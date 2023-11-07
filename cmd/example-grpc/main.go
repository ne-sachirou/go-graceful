// Example of gracefulgrpc.
package main

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/ne-sachirou/go-graceful"
	"github.com/ne-sachirou/go-graceful/gracefulgrpc"
)

func main() {
	ctx := context.Background()

	opts := []grpc.ServerOption{}
	srv := grpc.NewServer(opts...)
	reflection.Register(srv)

	if err := gracefulgrpc.ListenAndServe(
		ctx,
		":4317",
		srv,
		graceful.GracefulShutdownTimeout(time.Second),
	); err != nil {
		panic(err)
	}
}
