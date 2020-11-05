package main

import (
	"os"

	"github.com/pkg/errors"
	"github.com/segmentio/encoding/json"
)

type Config struct {
	DataSources     []DataSource
	DatabaseConfigs []DatabaseConfig
}

func readConfig(path string) (*Config, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file. path=%s", path)
	}
	defer configFile.Close()

	var config *Config
	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse file")
	}
	return config, nil
}
