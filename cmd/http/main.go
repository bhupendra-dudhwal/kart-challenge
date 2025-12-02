package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"kart-challenge/internal/builder"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	var envPath string
	flag.StringVar(&envPath, "env", "config/config.yaml", "Path to environment config file")
	flag.Parse()

	start := time.Now()
	ctx := context.Background()
	appBuilder := builder.NewAppBuilder(ctx)

	if err := appBuilder.LoadConfig(envPath); err != nil {
		log.Fatalf("failed to load config (%s): %+v", envPath, err)
	}

	logger, err := appBuilder.SetLogger()
	if err != nil {
		log.Fatalf("failed to initialize logger: %+v", err)
	}
	defer logger.Close()

	appBuilder.SetServices()
	appBuilder.SetHandlers()

	server, appConfig := appBuilder.Build()

	addr := fmt.Sprintf(":%d", appConfig.Port)

	logger.Info("Starting server", zap.String("addr", addr), zap.String("environment", appConfig.Environment.String()), zap.Int64("builder duration(Micro Sec)", time.Since(start).Microseconds()))
	go func() {
		if err := server.ListenAndServe(addr); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed to start", zap.Error(err))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	sig := <-stop
	logger.Info("Shutdown signal received", zap.String("signal", sig.String()))

	shutdownCtx, cancel := context.WithTimeoutCause(ctx, 10*time.Second, errors.New("server interrupt by os signal"))
	defer cancel()

	if err := server.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Info("graceful shutdown error", zap.Error(err))
	} else {
		logger.Info("server stopped gracefully")
	}
}
