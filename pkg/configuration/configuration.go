package configuration

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/xBlaz3kx/faceit-task/internal/repositories/mongo"
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

// GetConfiguration loads the configuration from Viper and validates it.
func GetConfiguration(viper *viper.Viper, configStruct interface{}) {
	// Load configuration from file
	err := viper.Unmarshal(configStruct)
	if err != nil {
		zap.L().Fatal("Cannot unmarshall", zap.Error(err))
	}

	// Validate configuration
	validationErr := validator.New().Struct(configStruct)
	if validationErr != nil {

		var errs validator.ValidationErrors
		errors.As(validationErr, &errs)

		hasErr := false
		for _, fieldError := range errs {
			zap.L().Error("Validation failed on field", zap.Error(fieldError))
			hasErr = true
		}

		if hasErr {
			zap.L().Fatal("Validation of the config failed")
		}
	}
}

// SetupEnv sets up the environment variables for the service.
func SetupEnv(cfgEngine *viper.Viper, serviceName string) {
	cfgEngine.SetEnvPrefix(serviceName)
	cfgEngine.AutomaticEnv()
	cfgEngine.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

// SetDefaults sets the default values for the app configuration.
func SetDefaults(cfgEngine *viper.Viper, serviceName string) {
	cfgEngine.SetDefault("database.uri", "")
	cfgEngine.SetDefault("server", ":8080")
}

// InitConfig initializes the configuration for the service.
func InitConfig(cfgEngine *viper.Viper, configurationFilePath string, additionalDirs ...string) {
	// Set defaults for searching for config files
	cfgEngine.SetConfigName("config")
	cfgEngine.SetConfigType("yaml")
	cfgEngine.AddConfigPath(".")
	cfgEngine.AddConfigPath("./config")

	for _, dir := range additionalDirs {
		cfgEngine.AddConfigPath(dir)
	}

	// Check if path is specified
	if configurationFilePath != "" {
		cfgEngine.SetConfigFile(configurationFilePath)
	}

	// Read the configuration
	err := cfgEngine.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			zap.L().With(zap.Error(err)).Warn("Config file not found")
		} else {
			zap.L().With(zap.Error(err)).Fatal("Something went wrong")
		}
	}
}
