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
	"strings"
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

	// Keep the judge service warm on Render free tier.
	// Render spins down a free service after 15 min of no incoming HTTP traffic.
	// The backend pings the judge /health every 10 minutes so it never sleeps.
	if cfg.Judge.Enabled && cfg.Judge.URL != "" {
		go keepJudgeWarm(cfg.Judge.URL, cfg.Judge.APIKey)
	}

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

// keepJudgeWarm pings the judge /health endpoint every 10 minutes so Render
// free-tier never spins it down (free services sleep after 15 min of no traffic).
func keepJudgeWarm(judgeURL, apiKey string) {
	baseURL := strings.TrimRight(judgeURL, "/")
	client := &http.Client{Timeout: 10 * time.Second}
	tick := time.NewTicker(10 * time.Minute)
	defer tick.Stop()
	for range tick.C {
		req, err := http.NewRequest(http.MethodGet, baseURL+"/health", nil)
		if err != nil {
			continue
		}
		if apiKey != "" {
			req.Header.Set("X-Judge-Key", apiKey)
		}
		resp, err := client.Do(req)
		if err != nil {
			logger.Warn("judge warm-ping failed", "error", err)
			continue
		}
		resp.Body.Close()
		logger.Info("judge warm-ping ok", "status", resp.StatusCode)
	}
}
