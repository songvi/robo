package generator

import (
	"github.com/songvi/robo/models"
)

type GeneratorConfig struct {
	Strategy        Strategy  `json:"strategy" yaml:"strategy"`
	FileStore       FileStore `json:"file_store" yaml:"file_store"`
	DBStore         DBStore   `json:"db_store" yaml:"db_store"`
	FileBuffer      int       `json:"file_buffer" yaml:"file_buffer"`
	UserBuffer      int       `json:"user_buffer" yaml:"user_buffer"`
	WorkspaceBuffer int       `json:"workspace_buffer" yaml:"workspace_buffer"`
	DBConfig        DBConfig  `json:"db_config" yaml:"db_config"`
}

// DBConfig holds the database configuration for GORM
type DBConfig struct {
	DSN string `json:"dsn" yaml:"dsn"` // Data Source Name for database connection
}
type Strategy struct {
	FileStrategy      models.FileStrategy      `json:"file_strategy" yaml:"file_strategy"`
	UserStrategy      models.UserStrategy      `json:"user_strategy" yaml:"user_strategy"`
	WorkspaceStrategy models.WorkspaceStrategy `json:"workspace_strategy" yaml:"workspace_strategy"`
}

type FileStore struct {
	FilePath string
}

type DBStore struct {
}
