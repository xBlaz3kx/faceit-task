package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xBlaz3kx/faceit-task/internal/api/grpc"
	"github.com/xBlaz3kx/faceit-task/internal/api/grpc/handlers"
	"github.com/xBlaz3kx/faceit-task/internal/api/http"
	"github.com/xBlaz3kx/faceit-task/internal/domain/services/user"
	"github.com/xBlaz3kx/faceit-task/internal/repositories/mongo"
	"github.com/xBlaz3kx/faceit-task/pkg/configuration"
	"github.com/xBlaz3kx/faceit-task/pkg/notifier"
	v1 "github.com/xBlaz3kx/faceit-task/pkg/proto/v1"
	"go.uber.org/zap"
)

const serviceName = "user"

var configFilePath string

var rootCmd = &cobra.Command{
	Use:     "user",
	Short:   "User service",
	Version: "0.1.0",
	Run: func(cmd *cobra.Command, args []string) {
		// Create a context that listens for the interrupt signal from the OS
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
		defer cancel()

		// Create a logger
		logger := zap.L()
		logger.Info("Starting the user service")

		// Get the application configuration
		cfg := configuration.AppConfig{}
		configuration.GetConfiguration(viper.GetViper(), &cfg)

		logger.Info("Loaded the configuration", zap.Any("config", cfg))

		// Connect to the database
		mongoHealthCheck := mongo.Connect(cfg.DatabaseCfg, logger)

		// Create the repository
		userRepository := mongo.NewUserRepository()

		// Create the user service
		userChangeNotifier := notifier.NewNotifier[user.ChangeStreamData]()
		userService := user.NewUserService(userRepository, userChangeNotifier)

		grpcServer := grpc.NewServer()

		// Register handler
		grpcUserHandler := handlers.NewUserGrpcHandler(userService)
		v1.RegisterUserServer(grpcServer, grpcUserHandler)

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
	},
}

func setupGlobalLogger() {
	// For simplicity, we'll just use the production logger
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)
}

func setupConfig() {
	cfgEngine := viper.GetViper()
	configuration.SetDefaults(cfgEngine, serviceName)
	configuration.SetupEnv(cfgEngine, serviceName)
	configuration.InitConfig(cfgEngine, configFilePath)
}

func main() {
	cobra.OnInitialize(setupGlobalLogger, setupConfig)

	// Add the flags to the root command
	rootCmd.PersistentFlags().StringVarP(&configFilePath, "config", "c", "", "Path to the configuration file")

	err := rootCmd.Execute()
	if err != nil {
		zap.L().Fatal("Failed to execute the root command", zap.Error(err))
	}
}
