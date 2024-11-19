package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	healthcheck "github.com/tavsec/gin-healthcheck"
	"github.com/tavsec/gin-healthcheck/checks"
	"github.com/tavsec/gin-healthcheck/config"
	"go.uber.org/zap"
)

type Server struct {
	Router *gin.Engine
	server *http.Server
	logger *zap.Logger
}

func NewServer(address string, logger *zap.Logger) *Server {
	// Create a Router and attach middleware
	router := gin.New()
	gin.SetMode(gin.ReleaseMode)

	return &Server{
		Router: router,
		logger: logger.Named("http-server"),
		server: &http.Server{
			Addr: address,
		},
	}
}

// Start starts the server with the provided healthchecks.
func (s *Server) Start(checks ...checks.Check) {
	s.logger.Info("Starting the server")

	// Attach recovery & log middleware
	s.Router.Use(ginzap.Ginzap(s.logger, time.RFC3339, true), ginzap.RecoveryWithZap(s.logger, true))

	// Add a healthcheck endpoint
	err := healthcheck.New(s.Router, config.DefaultConfig(), checks)
	if err != nil {
		s.logger.Panic("Cannot initialize healthcheck endpoint")
		return
	}

	s.server.Handler = s.Router.Handler()

	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatal("Failed to listen and serve", zap.Error(err))
		}
	}()
}

// Shutdown gracefully shuts down the server without interrupting any active connections.
func (s *Server) Shutdown() error {
	s.logger.Info("Shutting down the server")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}
