package main

import (
	"os"

	"github.com/pkg/errors"
	"github.com/segmentio/encoding/json"

	"github.com/minlau/mdb-tool/internal/utils/closer"
	"github.com/minlau/mdb-tool/store"
)

type Config struct {
	DataSources     []store.DataSource
	DatabaseConfigs []store.DatabaseConfig
}

func LoadConfig(path string) (*Config, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file. path=%s", path)
	}
	defer closer.Handle(configFile, "config file")

	var config *Config
	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse file")
	}
	return config, nil
}
