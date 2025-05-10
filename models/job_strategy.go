package models

type Strategy struct {
	CycleDuration int `json:"cycle_duration" yaml:"cycle_duration"`
	MaxUsers      int `json:"max_users" yaml:"max_users"`
	MaxFiles      int `json:"max_files" yaml:"max_files"`
	MaxWorkspaces int `json:"max_workspace" yaml:"max_workspace"`
}

type Cycle struct {
	UUID      string    `json:"uuid" yaml:"uuid" gorm:"primaryKey;type:uuid;"`
	Name      string    `json:"name" yaml:"name" gorm:"column:name;type:text;not null"`
	Strategy  *Strategy `json:"strategy" yaml:"strategy" gorm:"column:strategy;type:json"`
	StartedAt int64     `json:"started_at" yaml:"started_at" gorm:"column:started_at;type:bigint;not null"`
	DoneAt    int64     `json:"done_at" yaml:"done_at" gorm:"column:done_at;type:bigint"`
	Status    string    `json:"status" yaml:"status" gorm:"column:status;type:text;not null"`
}

type Session struct {
	UserID string `json:"user_id" yaml:"user_id"`
}

type JobServiceConfig struct {
	Strategy *Strategy `json:"job_strategy" yaml:"job_strategy"`
}
