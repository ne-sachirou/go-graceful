[![Go Reference](https://pkg.go.dev/badge/github.com/ne-sachirou/go-graceful.svg)](https://pkg.go.dev/github.com/ne-sachirou/go-graceful)
[![main branch](https://github.com/ne-sachirou/go-graceful/actions/workflows/on-push-main.yml/badge.svg)](https://github.com/ne-sachirou/go-graceful/actions/workflows/on-push-main.yml)

[English](#go-graceful) [日本語](#go-graceful-1)

# go-graceful

A library implemented in Go to gracefully shutdown the server.

## Functions

- Can start server gracefully.
  - Trap os.Signal and exit gracefully.
    - You can specify os.Signal to trap. By default, only os.Interrupt is trapped.
  - Exit processing can be specified.
    - If the specified timeout period elapses, a forced termination will occur. By default, it waits indefinitely for a normal termination.
- Multiple servers can be started gracefully.
  - If one of the servers fails to start, all started servers will be terminated gracefully.
- The HTTP and gRPC servers implement a wrapper function for easy startup.

## Usage Examples

### Start HTTP server gracefully

To start only one HTTP server, start the server with the `gracefulhttp.ListenAndServe` function.

[Example](cmd/example-http/main.go)

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ne-sachirou/go-graceful"
	"github.com/ne-sachirou/go-graceful/gracefulhttp"
)

func main() {
	ctx := context.Background()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("Hello, World!")); err != nil {
			fmt.Printf("failed to write response: %v\n", err)
		}
	})

	if err := gracefulhttp.ListenAndServe(
		ctx,
		":8000",
		mux,
		graceful.GracefulShutdownTimeout(time.Second),
	); err != nil {
		panic(err)
	}
}
```

For labstack/echo, pass the `*Echo` struct as a `net/http.Handler` interface to the `gracefulhttp.ListenAndServe` function instead of executing the `*Echo.Start` function.

### Start gRPC server gracefully

To start only one gRPC server, start the server with the `gracefulhttp.ListenAndServe` function.

[Example](cmd/example-grpc/main.go)

```go
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
```

### Start any server

Implement the `graceful.Server` interface and execute the `graceful.Servers.Graceful` function.

```go
package main

import (
	"context"
	"time"

	"github.com/ne-sachirou/go-graceful"
)

type ExampleBlockingServer struct {
	done       chan struct{}
	shutdowned chan struct{}
}

func (s *ExampleBlockingServer) Serve() error {
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
	close(s.done)
	select {
	case <-s.shutdowned:
		fmt.Println("example blocking shutdown")
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

	srv := graceful.Servers{Servers: []graceful.Server{&ExampleBlockingServer{}}}

	if err := srv.Graceful(ctx, graceful.GracefulShutdownTimeout(time.Second)); err != nil {
		panic(err)
	}
}
```

### Start multiple servers

Implement the `graceful.Server` interface and execute the `graceful.Servers.Graceful` function.

[Example](cmd/example/main.go)

```go
package main

import (
	"context"
	"time"

	"github.com/ne-sachirou/go-graceful"
)

// Example of main loop in go routine.
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
	return nil
}

func (s *ExampleServer) Shutdown(ctx context.Context) error {
	close(s.done)
	select {
	case <-s.shutdowned:
		fmt.Println("example shutdown")
		return nil
	case <-ctx.Done():
		if err := ctx.Err(); errors.Is(err, context.DeadlineExceeded) {
			return errors.Join(errors.New("failed to shutdown example"), context.DeadlineExceeded)
		}
		return context.Canceled
	}
}

// Example of main loop with caller blocking.
type ExampleBlockingServer struct {
	done       chan struct{}
	shutdowned chan struct{}
}

func (s *ExampleBlockingServer) Serve() error {
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
	close(s.done)
	select {
	case <-s.shutdowned:
		fmt.Println("example blocking shutdown")
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
```

# go-graceful

Go で実装した server を graceful に shutdown する library です。

## 機能

- server を graceful に起動できます
  - os.Signal を trap し、graceful に終了します
    - trap する os.Signal を指定できます。標準では os.Interrupt のみを trap します
  - 終了処理を指定できます
    - 指定した timeout 時間を過ぎると、強制終了します。標準では正常終了を永遠に待ちます
- 複数の server を graceful に起動できます
  - その内の server の 1 つでも起動に失敗したら、起動済みの全ての server を graceful に終了します
- HTTP と gRPC の server は、簡単に起動できる wrapper 函数を実装してあります

## 使用例

### HTTP server を graceful に起動する

HTTP server を 1 つだけ起動する場合は、`gracefulhttp.ListenAndServe` 函数で server を起動します。

[例](cmd/example-http/main.go)

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ne-sachirou/go-graceful"
	"github.com/ne-sachirou/go-graceful/gracefulhttp"
)

func main() {
	ctx := context.Background()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("Hello, World!")); err != nil {
			fmt.Printf("failed to write response: %v\n", err)
		}
	})

	if err := gracefulhttp.ListenAndServe(
		ctx,
		":8000",
		mux,
		graceful.GracefulShutdownTimeout(time.Second),
	); err != nil {
		panic(err)
	}
}
```

labstack/echo であれば、`*Echo.Start` 函数を実行する代はりに、`*Echo` struct を `net/http.Handler` interface として `gracefulhttp.ListenAndServe` 函数に渡してください。

### gRPC server を graceful に起動する

gRPC server を 1 つだけ起動する場合は、`gracefulgrpc.ListenAndServe` 函数で server を起動します。

[例](cmd/example-grpc/main.go)

```go
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
```

### 任意の server を起動する

`graceful.Server` interface を実装し、`graceful.Servers.Graceful` 函数を実行します。

```go
package main

import (
	"context"
	"time"

	"github.com/ne-sachirou/go-graceful"
)

type ExampleBlockingServer struct {
	done       chan struct{}
	shutdowned chan struct{}
}

func (s *ExampleBlockingServer) Serve() error {
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
	close(s.done)
	select {
	case <-s.shutdowned:
		fmt.Println("example blocking shutdown")
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

	srv := graceful.Servers{Servers: []graceful.Server{&ExampleBlockingServer{}}}

	if err := srv.Graceful(ctx, graceful.GracefulShutdownTimeout(time.Second)); err != nil {
		panic(err)
	}
}
```

### 複数の server を起動する

`graceful.Server` interface を実装し、`graceful.Servers.Graceful` 函数を実行します。

[例](cmd/example/main.go)

```go
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
	return nil
}

func (s *ExampleServer) Shutdown(ctx context.Context) error {
	close(s.done)
	select {
	case <-s.shutdowned:
		fmt.Println("example shutdown")
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
```
