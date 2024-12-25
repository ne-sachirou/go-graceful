// Package gracefulhttp provides utilities for go-graceful.
package gracefulhttp

import (
	"context"
	"errors"
	"net/http"

	"github.com/ne-sachirou/go-graceful"
)

type Server struct {
	Server *http.Server
}

func (s *Server) Serve() error {
	if err := s.Server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}

// ListenAndServe is a helper function like `http.ListenAndServe` for graceful.Servers.Graceful.
func ListenAndServe(
	ctx context.Context,
	addr string,
	handler http.Handler,
	options ...graceful.Option,
) error {
	httpSrv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	srv := graceful.Servers{Servers: []graceful.Server{&Server{Server: httpSrv}}}
	return srv.Graceful(ctx, options...)
}
