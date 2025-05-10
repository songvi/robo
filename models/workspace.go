package models

type Workspace struct {
	UUID      string   `json:"uuid" yaml:"uuid" gorm:"primaryKey;type:uuid;"`
	Name      string   `json:"name" yaml:"name" gorm:"column:name;type:text;not null"`
	Users     []string `json:"users" yaml:"users" gorm:"column:users;type:text;serializer:json;default:'[]'"`
	CycleID   string   `json:"cycle_id" yaml:"cycle_id" gorm:"column:cycle_id;type:uuid;not null"`
	SessionID string   `json:"session_id" yaml:"session_id" gorm:"column:session_id;type:text;not null"`
	// Foreign key relationships
	Cycle Cycle `gorm:"foreignKey:CycleID;references:UUID"`
}
