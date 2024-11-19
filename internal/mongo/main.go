package mongo

import (
	"sync"

	"github.com/kamva/mgm/v3"
	"github.com/tavsec/gin-healthcheck/checks"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.uber.org/zap"
)

var once sync.Once

type Configuration struct {
	URI string `yaml:"uri" json:"uri" mapstructure:"uri"`
}

// Connect connects to the database depending on the type.
func Connect(configuration Configuration, logger *zap.Logger) checks.Check {
	logger.Info("Attempting to connect to the database", zap.Any("configuration", configuration))

	// Connect to the database only once.
	once.Do(func() {
		clientOpts := options.Client()
		clientOpts.ApplyURI(configuration.URI)
		clientOpts.SetReadConcern(readconcern.Majority())

		err := mgm.SetDefaultConfig(nil, "faceit", clientOpts)
		if err != nil {
			logger.With(zap.Error(err)).Fatal("Unable to connect to database")
		}

		logger.Info("Connected to the database")
	})

	_, mongoClient, _, _ := mgm.DefaultConfigs()
	return checks.NewMongoCheck(10, mongoClient)
}
