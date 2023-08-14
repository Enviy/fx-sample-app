package config

import (
	"path/filepath"

	"go.uber.org/config"
	"go.uber.org/fx"
)

const (
	baseFile = "./base.yaml"
)

var Module = fx.Module(
	"config",
	fx.Provide(
		LoadConfig,
	),
)

// LoadConfig .
func LoadConfig() (config.Provider, error) {
	// Expand used for collecting env vars.
	// var lookup config.LookupFunc = func(string) (string, error) {
	//     return os.LookupEnv(key)
	// }
	// expandOpts := config.Expand(lookup)

	// local path refers to working dir of main \
	// at root of repo.
	cwd, err := filepath.Abs("./config")
	if err != nil {
		return nil, err
	}

	fileOpts := config.File(cwd + "/" + baseFile)
	cfg, err := config.NewYAML(fileOpts)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
