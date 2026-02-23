package config

import (
	"di0build/pkg/logger"
	"os"

	"gopkg.in/yaml.v3"
)

type SymLink struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

type Config struct {
	Symlinks []*SymLink `yaml:"links"`
	Packages []string   `yaml:"packages"`
}

func MustLoad(configPath string) *Config {
	var cfg Config

	file, err := os.ReadFile(configPath)
	if err != nil {
		logger.Fatal("Unable to read config: %v", err)
	}

	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		logger.Fatal("Unable to unmarshal config: %v", err)
	}

	return &cfg
}
