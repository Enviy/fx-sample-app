package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"go.uber.org/config"
	"go.uber.org/fx"
)

const (
	configDir = "./config"
	baseFile  = "/base.yaml"
	// devFile = "/dev.yaml"
)

var Module = fx.Module(
	"config",
	fx.Provide(
		Load,
	),
)

// Load service configs and default aws configs.
func Load() (config.Provider, aws.Config, error) {
	var acfg aws.Config
	var lookup config.LookupFunc = func(key string) (string, bool) {
		return os.LookupEnv(key)
	}
	expandOpts := config.Expand(lookup)

	cwd, err := filepath.Abs(configDir)
	if err != nil {
		return nil, acfg, fmt.Errorf("filepath abs %w", err)
	}

	fileOpts := config.File(cwd + baseFile)
	var ymlOpts []config.YAMLOption
	ymlOpts = append(ymlOpts, fileOpts)
	ymlOpts = append(ymlOpts, expandOpts)

	cfg, err := config.NewYAML(ymlOpts...)
	if err != nil {
		return nil, acfg, fmt.Errorf("config newyaml %w", err)
	}

	acfg, err = awsConfig.LoadDefaultConfig(
		context.Background(),
		awsConfig.WithRegion(cfg.Get("aws.region").String()),
	)
	if err != nil {
		return nil, acfg, fmt.Errorf("aws default config %w", err)
	}

	return cfg, acfg, nil
}
