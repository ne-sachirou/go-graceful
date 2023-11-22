// Package gracefulgrpc provides utilities for go-graceful.
package gracefulgrpc

import (
	"context"
	"errors"
	"net"

	"google.golang.org/grpc"

	"github.com/ne-sachirou/go-graceful"
)

type Server struct {
	Addr   string
	Server *grpc.Server
}

func (s *Server) Serve(ctx context.Context) error {
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	return s.Server.Serve(l)
}

func (s *Server) Shutdown(ctx context.Context) error {
	stopped := make(chan struct{})
	go func() {
		s.Server.GracefulStop()
		close(stopped)
	}()
	select {
	case <-stopped:
		return nil
	case <-ctx.Done():
		return errors.New("timeout")
	}
}

// ListenAndServe is a helper function for graceful.Servers.Graceful.
func ListenAndServe(
	ctx context.Context,
	addr string,
	server *grpc.Server,
	options ...graceful.Option,
) error {
	srv := graceful.Servers{Servers: []graceful.Server{&Server{Addr: addr, Server: server}}}
	return srv.Graceful(ctx, options...)
}
