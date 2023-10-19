package controller

import (
	"context"
	"fmt"
	"time"

	"fx-sample-app/gateway/cats"
	"fx-sample-app/repository/cache"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Controller .
type Controller interface {
	CatFact() (string, error)
}

type con struct {
	cat   *cats.Gateway
	log   *zap.Logger
	cache *cache.Cache
	keys  []string
}

type Params struct {
	fx.In

	Cat   *cats.Gateway
	Cache *cache.Cache
	Log   *zap.Logger
	Lc    fx.Lifecycle
}

// New .
func New(p Params) Controller {
	newController := &con{
		cat:   p.Cat,
		log:   p.Log,
		cache: p.Cache,
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
func (c *con) CatFact() (string, error) {
	fact, err := c.cat.GetFact()
	if err != nil {
		return "", err
	}

	//c.log.Info(fact)

	key := fmt.Sprintf("cat:%v", time.Now())
	c.keys = append(c.keys, key)
	c.cache.Set(key, fact, 5*time.Second)

	return fact, nil
}

func (c *con) listener(exitCh chan bool) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-exitCh:
			c.log.Info("closing goroutine")
			return
		case t := <-ticker.C:
			c.log.Info("ticker:", zap.Any("time", t))
			for _, key := range c.keys {
				value, exists := c.cache.Get(key)
				c.log.Info("cat record from cache",
					zap.Any("fact", value),
					zap.Bool("exists", exists),
				)
			}
		}
	}
}
