package config

import (
	"context"
	"encoding/json"
	"os"

	"go.uber.org/fx"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/songvi/robo/generator"
	"github.com/songvi/robo/logger"
)

// Config defines the application configuration
type Config struct {
	Broker      string                    `json:"broker"`
	Generator   generator.GeneratorConfig `json:"generator"`
	DSN         string                    `json:"dsn"`
	JobStrategy map[string]interface{}    `json:"job_strategy"`
}

// ConfigService defines the interface for configuration management
type ConfigService interface {
	GetConfig() Config
}

// configServiceImpl implements ConfigService
type configServiceImpl struct {
	config Config
}

// GetConfig returns the loaded configuration
func (c *configServiceImpl) GetConfig() Config {
	return c.config
}

// NewConfigService creates a new ConfigService instance
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

func NewGeneratorConfig(cfg ConfigService, logger logger.Logger) (generator.GeneratorConfig, error) {
	return cfg.GetConfig().Generator, nil
}

// Module defines the Fx module for ConfigService and GORM DB
var Module = fx.Module(
	"config",
	fx.Provide(NewGeneratorConfig),
	fx.Provide(NewConfigService),
	fx.Provide(func(lc fx.Lifecycle, configSvc ConfigService, logger logger.Logger) (*gorm.DB, error) {
		ctx := context.Background()
		cfg := configSvc.GetConfig()
		db, err := gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{})
		if err != nil {
			logger.Error(ctx, "Failed to open GORM database connection", "dsn", cfg.DSN, "error", err)
			return nil, err
		}

		lc.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				logger.Info(ctx, "Closing GORM database connection")
				sqlDB, err := db.DB()
				if err != nil {
					logger.Error(ctx, "Failed to get SQL DB from GORM", "error", err)
					return err
				}
				return sqlDB.Close()
			},
		})

		logger.Info(ctx, "GORM database connection established", "dsn", cfg.DSN)
		return db, nil
	}),
)
