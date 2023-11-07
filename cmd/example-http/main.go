// Example of gracefulhttp.
package main

import (
	"context"
	"net/http"
	"time"

	"github.com/ne-sachirou/go-graceful"
	"github.com/ne-sachirou/go-graceful/gracefulhttp"
)

func main() {
	ctx := context.Background()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("Hello, World!")) })

	gracefulhttp.ListenAndServe(ctx, ":8000", mux, graceful.GracefulShutdownTimeout(time.Second))
}
