package config

import (
	"context"
	"encoding/json"
	"os"

	"go.uber.org/fx"

	"github.com/songvi/robo/logger"
)

type Config struct {
	Broker    string                 `json:"broker"`
	Generator map[string]interface{} `json:"generator"`
	DSN       string                 `json:"dsn"`
}

type ConfigService interface {
	GetConfig() Config
}

type configServiceImpl struct {
	config Config
}

func (c *configServiceImpl) GetConfig() Config {
	return c.config
}

func NewConfigService(logger logger.Logger) (ConfigService, error) {
	ctx := context.Background()
	logger.Debug(ctx, "Loading config from config.json")
	file, err := os.Open("config.json")
	if err != nil {
		logger.Error(ctx, "Failed to open config.json", "error", err)
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		logger.Error(ctx, "Failed to decode config.json", "error", err)
		return nil, err
	}

	logger.Info(ctx, "Config loaded successfully", "broker", config.Broker)
	return &configServiceImpl{config: config}, nil
}

func ProvideConfigService() fx.Option {
	return fx.Provide(NewConfigService)
}
