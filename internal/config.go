package internal

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

const (
	configFilename = "config.yaml"
)

type Config struct {
	Host     string        `yaml:"host"`
	LogLevel zerolog.Level `yaml:"logLevel"`
}

func LoadConfig() (*Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(wd, configFilename)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
