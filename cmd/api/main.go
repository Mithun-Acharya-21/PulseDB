package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"myproject/internal/checker"
	"myproject/internal/db"
	"myproject/internal/metrics"
	"myproject/internal/monitor"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:secret@localhost:5432/postgres"
	}

	var repo monitor.Repository
	pool, err := db.New(ctx, connStr)
	if err != nil {
		slog.Warn("main: postgres not available, falling back to memory repo", "err", err)
		repo = monitor.NewMemoryRepo()
	} else {
		defer pool.Close()
		slog.Info("main: connected to postgres")
		repo = monitor.NewPostgresRepo(pool)
	}

	h := monitor.NewHandler(repo)
	router := gin.Default()
	router.Use(metrics.GinPrometheus()) 
	api := router.Group("/api/v1")
	{
		api.POST("/monitors", h.Create)               
		api.GET("/monitors", h.List)                  
		api.GET("/monitors/:id", h.GetByID)           
		api.PUT("/monitors/:id", h.Update)            
		api.DELETE("/monitors/:id", h.Delete)         
		api.GET("/monitors/:id/checks", h.ListChecks) 
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	chk := checker.New(repo)
	go chk.Start(ctx)

	// Build the HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		slog.Info("main: server listening", "addr", ":8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("main: server crashed", "err", err)
			cancel() // propagate failure to all goroutines
		}
	}()

	<-ctx.Done()
	slog.Info("main: shutdown signal received")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("main: graceful shutdown failed", "err", err)
	}

	slog.Info("main: stopped cleanly")
}
