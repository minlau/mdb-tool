package main

import (
	"github.com/pkg/errors"
	"github.com/segmentio/encoding/json"
	"os"
)

type Config struct {
	DataSources     []DataSource
	DatabaseConfigs []DatabaseConfig
}

func readConfig(path string) (*Config, error) {
	var config *Config

	configFile, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file. path=%s", path)
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		return nil, errors.Wrap(err, "failed to parse file")
	}
	return config, nil
}
