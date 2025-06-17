package server

import (
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/logger"
	"Weather-API-Application/internal/middleware"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Router *gin.Engine
	cfg    *config.Config
}

func NewServer(cfg *config.Config) *Server {
	router := gin.New()
	router.Use(middleware.Logger())
	router.Use(gin.Recovery())

	return &Server{
		cfg:    cfg,
		Router: router,
	}
}

func (s *Server) Run(ctx context.Context) {

	server := &http.Server{
		Addr:    s.cfg.AppPort,
		Handler: s.Router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(ctx, fmt.Errorf("listen: %s\n", err))
		}
	}()

	waitForSignal(ctx, server)
}

func waitForSignal(ctx context.Context, srv *http.Server) {

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info(ctx, "Shutdown Server ...")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal(ctx, fmt.Errorf("Server forced to shutdown: %w", err))
	}

	logger.Info(ctx, "Server exiting")
}
