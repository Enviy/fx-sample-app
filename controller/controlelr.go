package controller

import (
	"fx-sample-app/gateway/cats"

	"go.uber.org/zap"
)

// Interface .
type Interface interface {
	CatFact() (string, error)
}

type ctlr struct {
	cat    *cats.Gateway
	logger *zap.Logger
}

// New .
func New(c *cats.Gateway, l *zap.Logger) Interface {
	return &ctlr{
		cat:    c,
		logger: l,
	}
}

// CatWorkflow .
func (c *ctlr) CatFact() (string, error) {
	fact, err := c.cat.GetFact()
	if err != nil {
		return "", err
	}

	c.logger.Info(fact)

	return fact, nil
}
