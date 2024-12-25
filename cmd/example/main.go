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

func (s *ExampleServer) Serve() error {
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
	//return errors.New("faild to serve the server")
	return nil
}

func (s *ExampleServer) Shutdown(ctx context.Context) error {
	close(s.done)
	select {
	case <-s.shutdowned:
		fmt.Println("example shutdown")
		//time.Sleep(time.Second * 2)
		//return errors.New("faild to shutdown the server")
		return nil
	case <-ctx.Done():
		if err := ctx.Err(); errors.Is(err, context.DeadlineExceeded) {
			return errors.Join(errors.New("failed to shutdown example"), context.DeadlineExceeded)
		}
		return context.Canceled
	}
}

// 呼び出し元を blocking して main loop を回す例。
type ExampleBlockingServer struct {
	done       chan struct{}
	shutdowned chan struct{}
}

func (s *ExampleBlockingServer) Serve() error {
	s.done = make(chan struct{})
	s.shutdowned = make(chan struct{})
	fmt.Println("example blocking serve")
	//return errors.New("faild to serve the blocking server")
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
	close(s.done)
	select {
	case <-s.shutdowned:
		fmt.Println("example blocking shutdown")
		//time.Sleep(time.Second * 2)
		//return errors.New("faild to shutdown the blocking server")
		return nil
	case <-ctx.Done():
		if err := ctx.Err(); errors.Is(err, context.DeadlineExceeded) {
			return errors.Join(errors.New("failed to shutdown blocking example"), context.DeadlineExceeded)
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
