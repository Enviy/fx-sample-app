package cats

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/config"
)

// Gateway defines what methods have access to.
type Gateway struct {
	client *http.Client
	url    string
}

// New constructs a new gateway.
func New(cfg config.Provider) *Gateway {
	url := cfg.Get("catfact.url").String()
	return &Gateway{
		client: &http.Client{},
		url:    url,
	}
}

// GetFact calls the cat facts API.
func (g *Gateway) GetFact() (string, error) {
	resp, err := g.client.Get(g.url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var respObj RespObj
	err = json.Unmarshal(bodyBytes, &respObj)
	if err != nil {
		return "", err
	}

	if respObj.Fact == "" {
		return "", fmt.Errorf("empty fact")
	}

	return respObj.Fact, nil
}
