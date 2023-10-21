package redis

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// Gateway defines redis interaction methods.
type Gateway interface {
	Set(ctx context.Context, key, value string, exp time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

type gateway struct {
	client *redis.Client
}

// New is the redis gateway constructor.
func New() Gateway {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set.
		DB:       0,  // uses default db.
	})
	return &gateway{
		client: client,
	}
}

// Set stores a key value pair in a cache.
func (g *gateway) Set(ctx context.Context, key, value string, exp time.Duration) error {
	return g.client.Set(ctx, key, value, exp).Err()
}

// Get collects value by key from cache.
func (g *gateway) Get(ctx context.Context, key string) (string, error) {
	return g.client.Get(ctx, key).Result()
}
