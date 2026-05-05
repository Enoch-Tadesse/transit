package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type Pool struct {
	*pgxpool.Pool
}

// NewPool creates a pgxpool connection pool, verifies connectivity with a
// ping, and returns a ready to use Pool wrapper.
func NewPool(ctx context.Context, databaseURL string, logger zerolog.Logger) (*Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing postgres config: %w", err)
	}

	// moderate pool sizing for dev, tune these for production load
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConnIdleTime = 1 * time.Minute
	config.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("creating postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}

	logger.Info().Msg("connected to postgres")
	return &Pool{Pool: pool}, nil
}

func (p *Pool) Health(ctx context.Context) error {
	return p.Pool.Ping(ctx)
}
