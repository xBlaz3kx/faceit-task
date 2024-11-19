package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xBlaz3kx/faceit-task/internal/entrypoint"
	"github.com/xBlaz3kx/faceit-task/internal/pkg/configuration"
	"go.uber.org/zap"
)

const serviceName = "user"

var configFilePath string

var rootCmd = &cobra.Command{
	Use:     "user",
	Short:   "User service",
	Version: "0.1.0",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		// Create a logger
		logger := zap.L()
		logger.Info("Starting the user service")

		// Get the application configuration
		cfg := entrypoint.AppConfig{}
		configuration.GetConfiguration(viper.GetViper(), &cfg)

		logger.Info("Loaded the configuration", zap.Any("config", cfg))

		entrypoint.Run(ctx, cfg)
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
	// Create a context that listens for the interrupt signal from the OS
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer cancel()

	cobra.OnInitialize(setupGlobalLogger, setupConfig)

	// Add the flags to the root command
	rootCmd.PersistentFlags().StringVarP(&configFilePath, "config", "c", "", "Path to the configuration file")

	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		zap.L().Fatal("Failed to execute the root command", zap.Error(err))
	}
}
