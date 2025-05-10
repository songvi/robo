package models

type Workspace struct {
	// The name of the workspace
	Name  string   `json:"name" yaml:"name"`
	Users []string `json:"users" yaml:"users"`

	CycleID   string `json:"cycle_id" yaml:"cycle_id"`
	SessionID string `json:"session_id" yaml:"session_id"`
}
