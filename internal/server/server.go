package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/braden0236/o11y-golang/internal/metric"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine *gin.Engine
	http   *http.Server
}

func NewServer(m *metric.Metrics) *Server {

	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

    r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
        SkipPaths: []string{"/healthz", "/metrics"},
    }))

	r.Use(gin.Recovery())

	r.Use(m.Middleware())

	r.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "route not found"})
	})

	r.NoMethod(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "method not found"})
	})

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "healthy"})
	})

	r.GET("/metrics", m.Handler())

	return &Server{
		engine: r,
	}
}

func (s *Server) Run(port int) error {
	s.http = &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: s.engine,
	}
	log.Printf("Starting server on port %d", port)

	return s.http.ListenAndServe()
}

func (s *Server) WaitForShutdown(cleanup func()) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	log.Print("Shutdown signal received. Stopping server...")
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if cleanup != nil {
		cleanup()
	}

	if err := s.Shutdown(ctx); err != nil {
		log.Printf("Shutdown failed: %v", err)
		return err
	}

	log.Print("Server stopped gracefully.")
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.http == nil {
		return nil
	}
	return s.http.Shutdown(ctx)
}
