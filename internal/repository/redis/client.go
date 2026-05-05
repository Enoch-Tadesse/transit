package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Client struct {
	*goredis.Client
}

// NewClient creates a go-redis client, pings the server to verify
// connectivity, and returns a ready to use Client wrapper.
func NewClient(ctx context.Context, addr, password string, logger zerolog.Logger) (*Client, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("pinging redis: %w", err)
	}

	logger.Info().Str("addr", addr).Msg("connected to redis")
	return &Client{Client: client}, nil
}

func (c *Client) Health(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}
