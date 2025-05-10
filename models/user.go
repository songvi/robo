package models

// User represents a user with a display name, username, and language
type User struct {
	// The name of the user
	DisplayName string `json:"display_name" yaml:"display_name"`
	// Username
	UserName string `json:"username" yaml:"username"`
	// Language
	Language string `json:"language" yaml:"language"`

	UUID      string `json:"uuid" yaml:"uuid"`
	CycleID   string `json:"cycle_id" yaml:"cycle_id"`
	SessionID string `json:"session_id" yaml:"session_id"`
}
