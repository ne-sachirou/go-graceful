package grpc

import (
	"context"
	"errors"
	"net"

	"google.golang.org/grpc"
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
