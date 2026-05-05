package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/henok/transit-backend/internal/config"
	deliveryhttp "github.com/henok/transit-backend/internal/delivery/http"
	"github.com/henok/transit-backend/internal/repository/postgres"
	"github.com/henok/transit-backend/internal/repository/redis"
	"github.com/rs/zerolog"
)

// runMigrations applies any pending database schema changes on startup.
// uses golang-migrate with file based migrations from the /migrations directory.
func runMigrations(databaseURL string, logger zerolog.Logger) {
	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize migrations")
	}

	// ErrNoChange is fine, means schema is already up to date
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Fatal().Err(err).Msg("failed to run migrations")
	}

	logger.Info().Msg("migrations applied successfully")
}

// main is the application entry point. it loads config, runs migrations,
// connects to postgres and redis, starts the http server, and blocks until
// a SIGINT or SIGTERM signal triggers a graceful shutdown.
func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Logger()

	cfg := config.Load()

	runMigrations(cfg.DatabaseURL, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pg, err := postgres.NewPool(ctx, cfg.DatabaseURL, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer pg.Close()

	rdb, err := redis.NewClient(ctx, cfg.RedisAddr(), cfg.RedisPassword, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to redis")
	}
	defer rdb.Close()

	router := deliveryhttp.NewRouter(pg, rdb, logger)
	srv := deliveryhttp.NewServer(cfg.Port, router, logger)

	// run in background so we can block on the signal channel
	go func() {
		if err := srv.Start(); err != nil {
			logger.Fatal().Err(err).Msg("server failed")
		}
	}()

	// wait for kill signal before attempting graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("server forced to shutdown")
	}

	logger.Info().Msg("server exited gracefully")
}
