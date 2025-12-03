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

	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	for attempt := 1; attempt <= c.config.ConnectRetries; attempt++ {
		client = redis.NewClient(&redis.Options{
			Addr:         addr,
			Username:     c.config.Username,
			Password:     c.config.Password,
			DB:           c.config.Name,
			PoolSize:     c.config.PoolSize,
			MinIdleConns: c.config.MinIdleConns,
			DialTimeout:  c.config.DialTimeout,
			ReadTimeout:  c.config.ReadTimeout,
			WriteTimeout: c.config.WriteTimeout,
		})

		// Ping to verify connection
		ctxPing, cancel := context.WithTimeout(ctx, c.config.DialTimeout)
		defer cancel()
		if err = client.Ping(ctxPing).Err(); err != nil {
			c.logger.Error("Redis connection failed",
				zap.Int("attempt", attempt),
				zap.Int("maxAttempts", c.config.ConnectRetries),
				zap.String("addr", addr),
				zap.Error(err),
			)

			_ = client.Close()

			if attempt < c.config.ConnectRetries {
				backoff := time.Duration(1<<uint(attempt-1)) * c.config.RetryInterval
				if backoff > 30*time.Second {
					backoff = 30 * time.Second
				}

				c.logger.Warn("Retrying redis connection", zap.Duration("backoff", backoff))
				time.Sleep(backoff)
			}

			continue
		}

		c.logger.Info("Connected to Redis",
			zap.String("host", c.config.Host),
			zap.Int("port", c.config.Port),
			zap.Int("db", c.config.Name),
			zap.Int("poolSize", c.config.PoolSize),
		)

		c.client = client

		return client, nil
	}

	return nil, fmt.Errorf("failed to connect to Redis/KeyDB after %d attempts: %w", c.config.ConnectRetries, err)
}

func (c *cache) Close() error {
	if c.client == nil {
		return nil
	}

	return c.client.Close()
}
