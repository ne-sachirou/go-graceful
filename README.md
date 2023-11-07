# go-graceful

[![main branch](https://github.com/ne-sachirou/go-graceful/actions/workflows/on-push-main.yml/badge.svg)](https://github.com/ne-sachirou/go-graceful/actions/workflows/on-push-main.yml)

[日本語](#ja)

<h1 id="ja">go-graceful</h1>

Go で実装した server を graceful に shutdown する library です。

## 使用例

### http server を graceful に起動する

http server を 1 つだけ起動する場合は、`gracefulhttp.ListenAndServe` 函数で server を起動します。

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

### grpc server を graceful に起動する

GRPC server を 1 つだけ起動する場合は、`gracefulgrpc.ListenAndServe` 函数で server を起動します。

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

type ExampleBlockingServer struct{}

func (s *ExampleBlockingServer) Serve(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
		}
	}
}

func (s *ExampleBlockingServer) Shutdown(ctx context.Context) error {
	return nil
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
				return
			case <-time.After(time.Second):
			}
		}
	}()
	return nil
}

func (s *ExampleServer) Shutdown(ctx context.Context) error {
	return nil
}

// 呼び出し元を blocking して main loop を回す例。
type ExampleBlockingServer struct{}

func (s *ExampleBlockingServer) Serve(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
		}
	}
}

func (s *ExampleBlockingServer) Shutdown(ctx context.Context) error {
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
```
