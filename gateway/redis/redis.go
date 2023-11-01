package redis

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	redis "github.com/redis/go-redis/v9"
	"go.uber.org/config"
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
func New(cfg config.Provider, acfg aws.Config) Gateway {
	credCache := aws.NewCredentialsCache(acfg.Credentials)
	opts := &redis.Options{
		Addr:     cfg.Get("redis.address").String(),
		Password: "", // no password set.
		DB:       0,  // uses default db.
	}
	// If dev field is not empty, we're deployed, add provider.
	if cfg.Get("dev").String() != "" {
		opts.CredentialsProvider = func() (string, string) {
			creds, err := credCache.Retrieve(ctx)
			if err != nil {
				return "", ""
			}
			return "", creds.SessionToken
		}
	}
	client := redisNewClient(opts)
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
