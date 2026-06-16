// Package logger provides a thin leveled logger over the standard library slog.
package logger

import (
	"log/slog"
	"os"
)

var log *slog.Logger

// Init configures the global logger. Use JSON in production, text otherwise.
func Init(env string) {
	level := slog.LevelDebug
	if env == "production" {
		level = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: level}
	var h slog.Handler
	if env == "production" {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	log = slog.New(h)
	slog.SetDefault(log)
}

func L() *slog.Logger {
	if log == nil {
		Init("development")
	}
	return log
}

func Info(msg string, args ...any)  { L().Info(msg, args...) }
func Error(msg string, args ...any) { L().Error(msg, args...) }
func Warn(msg string, args ...any)  { L().Warn(msg, args...) }
func Debug(msg string, args ...any) { L().Debug(msg, args...) }
func Fatal(msg string, args ...any) {
	L().Error(msg, args...)
	os.Exit(1)
}
