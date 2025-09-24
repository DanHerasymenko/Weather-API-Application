package server

import (
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/logger"
	"Weather-API-Application/internal/middleware"
	"Weather-API-Application/internal/utils/response"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	Router *gin.Engine
	cfg    *config.Config
}

func NewServer(cfg *config.Config) *Server {
	router := gin.New()
	router.Use(middleware.Logger())
	router.Use(gin.Recovery())

	// Serve static assets
	router.Static("/static", "./static")
	router.GET("/", func(c *gin.Context) { c.File("./static/index.html") })

	// Swagger UI handler
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 404 handler to log and return a consistent error body
	router.NoRoute(func(c *gin.Context) {
		response.WriteErrorJSON(c, http.StatusNotFound, fmt.Errorf("route not found: %s %s", c.Request.Method, c.Request.URL.Path), "Route not found")
	})

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
