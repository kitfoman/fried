package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kitfoman/gpuFryer/pkg/server"
)

func main() {
	var (
		listenAddr = flag.String("addr", ":8080", "HTTP server address to listen on")
		logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	)
	flag.Parse()

	// Configure slog
	var level slog.Level
	switch *logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting gpuFryer", "logLevel", *logLevel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		slog.Info("Received signal", "signal", sig)
		cancel()
	}()

	srv := server.NewServer(*listenAddr, logger)
	if err := srv.Start(ctx); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}

	fmt.Println("Server shutdown complete")
}

