package models

type File struct {
	UUID          string `json:"uuid" yaml:"uuid" gorm:"primaryKey;type:uuid;"`
	Name          string `json:"name" yaml:"name" gorm:"column:name;type:text;not null"`
	CycleID       string `json:"cycle_id" yaml:"cycle_id" gorm:"column:cycle_id;type:uuid;not null"`
	SessionID     string `json:"session_id" yaml:"session_id" gorm:"column:session_id;type:text;not null"`
	Description   string `json:"description" yaml:"description" gorm:"column:description;type:text"`
	FileExtension string `json:"file_extension" yaml:"file_extension" gorm:"column:file_extension;type:text;not null"`
	FileSize      int    `json:"file_size" yaml:"file_size" gorm:"column:file_size;type:integer;not null"`
	FileContent   string `json:"file_content" yaml:"file_content" gorm:"column:file_content;type:text"`
	WorkspaceID   string `json:"workspace_id" yaml:"workspace_id" gorm:"column:workspace_id;type:uuid;not null"`
	// Foreign key relationships
	Cycle     Cycle     `gorm:"foreignKey:CycleID;references:UUID"`
	Workspace Workspace `gorm:"foreignKey:WorkspaceID;references:UUID"`
}
