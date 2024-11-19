package grpc

import (
	"net"
	"runtime/debug"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type Server struct {
	logger *zap.Logger
	server *grpc.Server
}

func NewServer() *Server {
	logger := zap.L().Named("grpc-server")

	// Create a GRPC server with recovery and logger interceptor
	recoveryHandler := func(p any) (err error) {
		logger.Error("recovered from panic", zap.Any("panic", p), zap.String("stack", string(debug.Stack())))
		return status.Errorf(codes.Internal, "%s", p)
	}

	server := grpc.NewServer(
		// Logger and recovery unary interceptors
		grpc.ChainUnaryInterceptor(
			grpc_zap.UnaryServerInterceptor(logger),
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(recoveryHandler)),
		),
		// Logger and recovery stream interceptors
		grpc.ChainStreamInterceptor(
			grpc_zap.StreamServerInterceptor(logger),
			grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandler(recoveryHandler)),
		),
	)

	// Register the healthcheck
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	return &Server{
		logger: logger,
		server: server,
	}
}

func (s *Server) RegisterService(service *grpc.ServiceDesc, impl interface{}) {
	s.server.RegisterService(service, impl)
}

// Start starts the gRPC server, listening on the provided address
func (s *Server) Start(address string) {
	s.logger.Info("Starting gRPC server")

	lis, err := net.Listen("tcp", address)
	if err != nil {
		s.logger.Fatal("failed to listen", zap.Error(err))
	}

	go func() {
		if err := s.server.Serve(lis); err != nil {
			s.logger.Fatal("failed to serve gRPC", zap.Error(err))
		}
	}()
}

// Stop gracefully shuts down the gRPC server
func (s *Server) Stop() {
	s.logger.Info("Gracefully shutting down gRPC server")
	s.server.GracefulStop()
}
