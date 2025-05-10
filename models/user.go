package models

type User struct {
	UUID        string `json:"uuid" yaml:"uuid" gorm:"primaryKey;type:uuid;"`
	DisplayName string `json:"display_name" yaml:"display_name" gorm:"column:display_name;type:text;not null"`
	UserName    string `json:"username" yaml:"username" gorm:"column:username;type:text;unique;not null"`
	Language    string `json:"language" yaml:"language" gorm:"column:language;type:text;not null"`
	CycleID     string `json:"cycle_id" yaml:"cycle_id" gorm:"column:cycle_id;type:uuid;not null"`
	SessionID   string `json:"session_id" yaml:"session_id" gorm:"column:session_id;type:text;not null"`
	// Foreign key relationships
	Cycle Cycle `gorm:"foreignKey:CycleID;references:UUID"`
}
