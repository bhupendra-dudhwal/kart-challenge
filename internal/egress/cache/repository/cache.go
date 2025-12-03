package repository

import (
	"context"
	"fmt"

	egressPorts "github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports/egress"
	"github.com/redis/go-redis/v9"
)

type cacheRepository struct {
	redisClient *redis.Client
}

func NewRepository(redisClient *redis.Client) egressPorts.CacheRepository {
	return &cacheRepository{
		redisClient: redisClient,
	}
}

func (c *cacheRepository) SAdd(ctx context.Context, key string, members ...interface{}) error {
	if err := c.redisClient.SAdd(ctx, key, members...).Err(); err != nil {
		return fmt.Errorf("redis SAdd error: %w", err)
	}
	return nil
}

func (c *cacheRepository) Do(ctx context.Context, args []interface{}) error {
	if err := c.redisClient.Do(ctx, args...).Err(); err != nil {
		return fmt.Errorf("redis do error: %w", err)
	}
	return nil
}

func (c *cacheRepository) BFExists(ctx context.Context, filterName string, value string) (bool, error) {
	res, err := c.redisClient.Do(ctx, "BF.EXISTS", filterName, value).Bool()
	if err != nil {
		return false, fmt.Errorf("redis BF.EXISTS error: %w", err)
	}
	return res, nil
}
