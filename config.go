package main

import (
	"encoding/json"
	"fmt"
	"os"

	"go.uber.org/fx"

	"github.com/songvi/robo/generator"
)

// Config represents the application configuration
type Config struct {
	Generator generator.GeneratorConfig `json:"generator" yaml:"generator"`
	Dsn       string                    `json:"dsn" yaml:"dsn"`
}

// ConfigService provides access to the application's configuration
type ConfigService interface {
	GetConfig() *Config
}

// configServiceImpl is the implementation of ConfigService
type configServiceImpl struct {
	config *Config
}

// NewConfigService initializes the ConfigService and loads the configuration from a file
func NewConfigService(configPath string) (ConfigService, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &configServiceImpl{config: &cfg}, nil
}

// GetConfig returns the loaded configuration
func (c *configServiceImpl) GetConfig() *Config {
	return c.config
}

// ProvideConfigService is an fx-compatible constructor for ConfigService
func ProvideConfigService() fx.Option {
	return fx.Provide(func() (ConfigService, error) {
		// Replace with your actual config file path
		const configPath = "config.json"
		return NewConfigService(configPath)
	})
}
