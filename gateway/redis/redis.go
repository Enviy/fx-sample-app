package redis

import (
	"context"

	redis "github.com/redis/go-redis/v9"
)

// Gateway defines redis interaction methods.
type Gateway interface {
	Set(ctx context.Context, key, value string) error
	Get(ctx cotext.Context, key string) (string, error)
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
	return &gateway{}
}

// Set stores a key value pair in a cache.
func Set(ctx context.Context, key, value string) error {
	return g.client.Set(ctx, key, value)
}

// Get collects value by key from cache.
func Get(ctx context.Context, key string) (string, error) {
	return g.client.Get(ctx, key).Result()
}
