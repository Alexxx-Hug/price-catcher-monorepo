package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/Alexxx-Hug/price-catcher-monorepo/gen/go/monitor"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	port    string
	server  *grpc.Server
	handler *MonitorHandler
	logger  *zap.Logger
}

func NewServer(port string, handler *MonitorHandler, logger *zap.Logger) *Server {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Server{
		port:    port,
		server:  grpc.NewServer(),
		handler: handler,
		logger:  logger,
	}
}

func (s *Server) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("listen grpc port %s: %w", s.port, err)
	}

	monitor.RegisterMonitorServiceServer(s.server, s.handler)
	reflection.Register(s.server)

	go func() {
		<-ctx.Done()
		s.logger.Info("grpc server stopping")
		s.server.GracefulStop()
	}()

	s.logger.Info("grpc server started", zap.String("port", s.port))

	if err := s.server.Serve(listener); err != nil {
		return fmt.Errorf("serve grpc: %w", err)
	}

	return nil
}
