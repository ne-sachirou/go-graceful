// Graceful shutdown example.
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ne-sachirou/go-graceful"
)

// main loop を go routine の中で回す例。
type ExampleServer struct {
	done       chan struct{}
	shutdowned chan struct{}
}

func (s *ExampleServer) Serve(ctx context.Context) error {
	s.done = make(chan struct{})
	s.shutdowned = make(chan struct{})
	go func() {
		for {
			select {
			case <-s.done:
				fmt.Println("example done")
				close(s.shutdowned)
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
	close(s.done)
	select {
	case <-s.shutdowned:
		return nil
	case <-ctx.Done():
		if err := ctx.Err(); errors.Is(err, context.DeadlineExceeded) {
			return context.DeadlineExceeded
		}
		return context.Canceled
	}
}

// 呼び出し元を blocking して main loop を回す例。
type ExampleBlockingServer struct {
	done       chan struct{}
	shutdowned chan struct{}
}

func (s *ExampleBlockingServer) Serve(ctx context.Context) error {
	s.done = make(chan struct{})
	s.shutdowned = make(chan struct{})
	fmt.Println("example blocking serve")
	for {
		select {
		case <-s.done:
			fmt.Println("example blocking done")
			close(s.shutdowned)
			return nil
		case <-time.After(time.Second):
			fmt.Println("example blocking working")
		}
	}
}

func (s *ExampleBlockingServer) Shutdown(ctx context.Context) error {
	fmt.Println("example blocking shutdown")
	close(s.done)
	select {
	case <-s.shutdowned:
		return nil
	case <-ctx.Done():
		if err := ctx.Err(); errors.Is(err, context.DeadlineExceeded) {
			return context.DeadlineExceeded
		}
		return context.Canceled
	}
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
