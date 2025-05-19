package server

import (
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/logger"
	"Weather-API-Application/internal/services"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	Router *gin.Engine
	cfg    *config.Config
	srvc   *services.Services
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg:    cfg,
		Router: gin.New(),
	}
}

// Run starts the HTTP server in a separate goroutine and handles graceful shutdown.
//
//   - Binds the server to the configured port.
//   - Uses the Gin router as the HTTP handler.
//   - Logs fatal error if the server fails unexpectedly.
//   - Listens for OS shutdown signals (SIGINT/SIGTERM) and performs graceful shutdown
//     using the waitForSignal helper.
func (s *Server) Run(ctx context.Context) {

	server := &http.Server{
		Addr:    s.cfg.AppPort,
		Handler: s.Router,
	}

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(ctx, fmt.Errorf("listen: %s\n", err))
		}
	}()

	// Graceful shutdown
	waitForSignal(ctx, server)
}

// waitForSignal helper function for graceful shutdown of the server.
func waitForSignal(ctx context.Context, srv *http.Server) {

	// Wait for interrupt signal to gracefully shut down the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)

	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info(ctx, "Shutdown Server ...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal(ctx, fmt.Errorf("Server forced to shutdown: %w", err))
	}

	logger.Info(ctx, "Server exiting")
}
