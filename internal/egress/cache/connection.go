package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/models"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports"
	egressPorts "github.com/bhupendra-dudhwal/kart-challenge/internal/core/ports/egress"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type cache struct {
	client *redis.Client
	config *models.Cache
	logger ports.LoggerPorts
}

func NewCache(config *models.Cache, logger ports.LoggerPorts) egressPorts.CacheConnectionPorts {
	return &cache{
		config: config,
		logger: logger,
	}
}

func (c *cache) Connect(ctx context.Context) (*redis.Client, error) {
	var (
		client *redis.Client
		err    error
	)

	// Retry logic
	for attempt := 1; attempt <= c.config.ConnectRetries; attempt++ {
		client = redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("%s:%d", c.config.Host, c.config.Port),
			// Username:     c.config.Username,
			// Password:     c.config.Password,
			DB:           c.config.Name,
			PoolSize:     c.config.PoolSize,
			MinIdleConns: c.config.MinIdleConns,
			DialTimeout:  c.config.DialTimeout,
			ReadTimeout:  c.config.ReadTimeout,
			WriteTimeout: c.config.WriteTimeout,
		})

		// Ping to verify connection
		err = client.Ping(ctx).Err()
		if err == nil {
			c.logger.Info("Connected to Redis/KeyDB",
				zap.String("host", c.config.Host),
				zap.Int("port", c.config.Port),
				zap.Int("db", c.config.Name),
				zap.Int("poolSize", c.config.PoolSize),
			)
			c.client = client
			return client, nil
		}

		c.logger.Error("Redis/KeyDB connection failed",
			zap.Int("attempt", attempt),
			zap.Int("maxAttempts", c.config.ConnectRetries),
			zap.Error(err),
		)

		if attempt < c.config.ConnectRetries {
			sleep := c.config.RetryInterval
			if sleep <= 0 {
				sleep = time.Second * time.Duration(attempt) // incremental backoff
			}
			time.Sleep(sleep)
		}
	}

	return nil, fmt.Errorf("failed to connect to Redis/KeyDB after %d attempts: %w", c.config.ConnectRetries, err)
}

func (c *cache) Close() error {
	if c.client == nil {
		return nil
	}

	return c.client.Close()
}
