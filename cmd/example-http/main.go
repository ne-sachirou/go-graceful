// Example of gracefulhttp.
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
