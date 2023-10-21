package controller

import (
	"context"
	"fmt"
	"time"

	"fx-sample-app/gateway/cats"
	"fx-sample-app/gateway/redis"
	"fx-sample-app/gateway/slack"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Controller .
type Controller interface {
	CatFact(ctx context.Context) (string, error)
}

type con struct {
	cat   *cats.Gateway
	log   *zap.Logger
	cache redis.Gateway
	slack slack.Gateway
	keys  []string
}

type Params struct {
	fx.In

	Cat   *cats.Gateway
	Cache redis.Gateway
	Slack slack.Gateway
	Log   *zap.Logger
	Lc    fx.Lifecycle
}

// New .
func New(p Params) Controller {
	newController := &con{
		cat:   p.Cat,
		log:   p.Log,
		cache: p.Cache,
		slack: p.Slack,
	}

	exitCh := make(chan bool, 1)
	p.Lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go newController.listener(exitCh)
			return nil
		},
		OnStop: func(context.Context) error {
			exitCh <- true
			return nil
		},
	})

	return newController
}

// CatWorkflow .
func (c *con) CatFact(ctx context.Context) (string, error) {
	fact, err := c.cat.GetFact()
	if err != nil {
		return "", err
	}

	//c.log.Info(fact)

	key := fmt.Sprintf("cat:%v", time.Now())
	c.keys = append(c.keys, key)
	c.cache.Set(ctx, key, fact, 1*time.Minute)

	return fact, nil
}

func (c *con) listener(exitCh chan bool) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-exitCh:
			c.log.Info("closing goroutine")
			return
		case t := <-ticker.C:
			ctx := context.Background()
			c.log.Info("ticker reached", zap.Any("time", t))

			// Instead of HScan, just ranging over slice for now.
			for _, key := range c.keys {
				value, err := c.cache.Get(ctx, key)
				if err != nil {
					if err.Error() == "redis: nil" {
						continue
					}
					// could emit err to channel here.
					c.log.Error("cache Get", zap.Error(err))
				}
				c.log.Info("cat record from cache",
					zap.Any("fact", value),
				)
			}
		}
	}
}
