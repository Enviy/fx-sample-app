package config

import (
	"os"
	"path/filepath"

	"go.uber.org/config"
	"go.uber.org/fx"
)

const (
	configDir = "./config"
	baseFile  = "/base.yaml"
	// devFile = "/dev.yaml"
	// secretsFile = "/secrets.yaml"
)

var Module = fx.Module(
	"config",
	fx.Provide(
		Load,
	),
)

// Load default configs.
func Load() (config.Provider, error) {
	// Expand used for collecting env vars.
	var lookup config.LookupFunc = func(string) (string, error) {
		return os.LookupEnv(key)
	}
	expandOpts := config.Expand(lookup)
	cwd, err := filepath.Abs(configDir)
	if err != nil {
		return nil, err
	}

	fileOpts := config.File(cwd + baseFile)
	var ymlOpts []config.YAMLOption
	ymlOpts = append(ymlOpts, fileOpts)
	ymlOpts = append(ymlOpts, expandOpts)

	cfg, err := config.NewYAML(ymlOpts...)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
