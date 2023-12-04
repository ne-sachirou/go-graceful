// Graceful shutdown example.
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ne-sachirou/go-graceful"
)

// main loop を go routine の中で回す例。
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

// 呼び出し元を blocking して main loop を回す例。
type ExampleBlockingServer struct{}

func (s *ExampleBlockingServer) Serve(ctx context.Context) error {
	fmt.Println("example blocking serve")
	for {
		select {
		case <-ctx.Done():
			fmt.Println("example blocking done")
			return context.Canceled
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

	if err := srv.Graceful(ctx, graceful.GracefulShutdownTimeout(time.Second)); err != nil {
		panic(err)
	}
}
