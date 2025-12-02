package egress

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type CacheConnectionPorts interface {
	Connect(ctx context.Context) (*redis.Client, error)
	Close() error
}

type CacheRepository interface {
	SAdd(ctx context.Context, key string, members ...interface{}) error
	Do(ctx context.Context, args []interface{}) error
}
