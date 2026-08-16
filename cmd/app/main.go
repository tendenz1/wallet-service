package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"wallet/internal/config"
	"wallet/internal/handler"
	"wallet/internal/repository"
	"wallet/internal/service"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("application stopped with error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	poolConfig, err := pgxpool.ParseConfig(cfg.Postgres.DSN())
	if err != nil {
		return fmt.Errorf("parse database config: %w", err)
	}
	poolConfig.MaxConns = cfg.Postgres.MaxConns
	connectCtx, cancelConnect := context.WithTimeout(context.Background(), cfg.Postgres.ConnectTimeout)
	defer cancelConnect()
	pool, err := pgxpool.NewWithConfig(connectCtx, poolConfig)
	if err != nil {
		return fmt.Errorf("create database pool: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(connectCtx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	logger.Info("connected to PostgreSQL", "host", cfg.Postgres.Host, "database", cfg.Postgres.Database)
	walletRepository := repository.NewPostgresRepository(pool)
	walletService := service.NewWalletService(walletRepository)
	walletHandler := handler.NewWalletHandler(walletService, logger)
	server := &http.Server{
		Addr:    cfg.Address(),
		Handler: walletHandler.Router(),
	}
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("wallet service started", "address", cfg.Address())
		serverErrors <- server.ListenAndServe()
	}()
	signalCtx, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()
	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	case <-signalCtx.Done():
		logger.Info("shutting down wallet service")
	}
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}
	if err := <-serverErrors; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP after shutdown: %w", err)
	}
	logger.Info("wallet service stopped")
	return nil
}
