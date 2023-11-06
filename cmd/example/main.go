package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ne-sachirou/go-graceful"
)

type ExampleServer struct{}

func (s *ExampleServer) Serve(ctx context.Context) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("example done")
				return
			case <-time.After(time.Second):
				fmt.Println("example working")
			}
		}
	}()
	fmt.Println("example serve")
	return nil
}

func (s *ExampleServer) Shutdown(ctx context.Context) error {
	fmt.Println("example shutdown")
	return nil
}

type ExampleBlockingServer struct{}

func (s *ExampleBlockingServer) Serve(ctx context.Context) error {
	fmt.Println("example blocking serve")
	for {
		select {
		case <-ctx.Done():
			fmt.Println("example blocking done")
			return nil
		case <-time.After(time.Second):
			fmt.Println("example blocking working")
		}
	}
}

func (s *ExampleBlockingServer) Shutdown(ctx context.Context) error {
	fmt.Println("example blocking shutdown")
	return nil
}

func main() {
	ctx := context.Background()
	srv := graceful.Servers{
		Servers: []graceful.Server{
			&ExampleServer{},
			&ExampleBlockingServer{},
		},
	}
	cfg := graceful.Config{ShutdownTimeout: time.Second}
	if err := srv.Graceful(ctx, cfg); err != nil {
		panic(err)
	}
}
