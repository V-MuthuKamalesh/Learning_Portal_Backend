// Command api is the entrypoint for the College Assessment Platform backend.
//
// @title College Assessment Platform API
// @version 1.0
// @description Multi-tenant MCQ & coding assessment platform.
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/collegeassess/backend/configs"
	"github.com/collegeassess/backend/internal/database"
	"github.com/collegeassess/backend/internal/server"
	"github.com/collegeassess/backend/pkg/dotenv"
	"github.com/collegeassess/backend/pkg/logger"
)

func main() {
	dotenv.Load(".env")
	cfg := configs.Load()
	logger.Init(cfg.App.Env)
	logger.Info("starting", "app", cfg.App.Name, "env", cfg.App.Env)

	db, err := database.NewPostgres(cfg)
	if err != nil {
		logger.Fatal("database connect failed", "error", err)
	}
	if err := database.Migrate(db); err != nil {
		logger.Fatal("migration failed", "error", err)
	}
	if err := database.Seed(db, cfg); err != nil {
		logger.Fatal("seed failed", "error", err)
	}

	rdb, err := database.NewRedis(cfg)
	if err != nil {
		logger.Warn("redis unavailable, continuing without cache", "error", err)
	}

	engine := server.New(server.Deps{Cfg: cfg, DB: db, Redis: rdb})

	srv := &http.Server{
		Addr:              ":" + cfg.App.Port,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("listening", "port", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server error", "error", err)
		}
	}()

	// Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("forced shutdown", "error", err)
	}
}
