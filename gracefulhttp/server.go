// Package gracefulhttp provides utilities for go-graceful.
package gracefulhttp

import (
	"context"
	"errors"
	"net/http"
)

type Server struct {
	Server *http.Server
}

func (s *Server) Serve(ctx context.Context) error {
	if err := s.Server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}
