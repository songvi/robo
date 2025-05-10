package models

import "encoding/json"

type Job struct {
	UUID       string          `json:"uuid" yaml:"uuid" gorm:"primaryKey;type:uuid;"`
	WorkerID   string          `json:"worker_id" yaml:"worker_id" gorm:"column:worker_id;type:uuid"`
	Name       string          `json:"name" yaml:"name" gorm:"column:name;type:text;not null"`
	InputData  json.RawMessage `json:"input_data" yaml:"input_data" gorm:"column:input_data;type:json"`
	OutputData json.RawMessage `json:"output_data" yaml:"output_data" gorm:"column:output_data;type:json"`
	Error      string          `json:"error" yaml:"error" gorm:"column:error;type:text"`
	StartAt    int64           `json:"start_at" yaml:"start_at" gorm:"column:start_at;type:bigint"`
	DoneAt     int64           `json:"done_at" yaml:"done_at" gorm:"column:done_at;type:bigint"`
	Status     string          `json:"status" yaml:"status" gorm:"column:status;type:text;not null"`
	CycleUUID  string          `json:"cycle_uuid" yaml:"cycle_uuid" gorm:"column:cycle_uuid;type:uuid;not null"`
	SessionID  string          `json:"session_id" yaml:"session_id" gorm:"column:session_id;type:text;not null"`
	// Foreign key relationships
	Cycle  Cycle  `gorm:"foreignKey:CycleUUID;references:UUID"`
	Worker Worker `gorm:"foreignKey:WorkerID;references:UUID"`
}
