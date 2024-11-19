package entrypoint

import (
	"context"

	"github.com/xBlaz3kx/faceit-task/internal/domain/users"
	grpc2 "github.com/xBlaz3kx/faceit-task/internal/grpc"
	"github.com/xBlaz3kx/faceit-task/internal/mongo"
	"github.com/xBlaz3kx/faceit-task/internal/pkg/grpc"
	"github.com/xBlaz3kx/faceit-task/internal/pkg/http"
	"go.uber.org/zap"
)

type AppConfig struct {
	// Server is the address the server will listen on
	Server string `yaml:"server" json:"server" mapstructure:"server"`

	// DatabaseCfg contains the connection URI for the database
	DatabaseCfg mongo.Configuration `yaml:"database" json:"database" mapstructure:"database"`

	// todo possible improvements:
	// - add observability configuration (logs + tracing)
}

func Run(ctx context.Context, cfg AppConfig) {
	// Create a logger
	logger := zap.L()
	logger.Info("Starting the user service", zap.Any("configuration", cfg))

	// Connect to the database
	mongoHealthCheck := mongo.Connect(cfg.DatabaseCfg, logger)

	// Create the repository
	userRepository := mongo.NewUserRepository()

	// Create the user service
	userService := users.NewUserService(userRepository)

	grpcServer := grpc.NewServer()

	// Register handler
	grpcUserHandler := grpc2.NewUserGrpcHandler(userService)
	grpc2.RegisterUserServer(grpcServer, grpcUserHandler)

	grpcServer.Start(cfg.Server)

	// Create and start the HTTP server
	httpServer := http.NewServer(":80", logger)
	httpServer.Start(mongoHealthCheck)

	// Wait for the interrupt signal
	<-ctx.Done()

	// Gracefully shutdown the GRPC server
	grpcServer.Stop()

	// Shutdown the HTTP server
	err := httpServer.Shutdown()
	if err != nil {
		logger.Fatal("Failed to shutdown the HTTP server", zap.Error(err))
	}
}
