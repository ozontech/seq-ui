package redisclient

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	config "github.com/ozontech/seq-ui/internal/app/config/v2"
)

func New(ctx context.Context, cfg *config.Redis) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Username: cfg.Username,
		Password: cfg.Password,

		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,

		MaxRetries:      cfg.MaxRetries,
		MinRetryBackoff: cfg.MinRetryBackoff,
		MaxRetryBackoff: cfg.MaxRetryBackoff,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}
