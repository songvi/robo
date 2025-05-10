package dispatcher

type Strategy struct {
	CycleDuration int `json:"cycle_duration" yaml:"cycle_duration"`
	MaxUsers      int `json:"max_users" yaml:"max_users"`
	MaxFiles      int `json:"max_files" yaml:"max_files"`
	MaxWorkspaces int `json:"max_workspace" yaml:"max_workspace"`
}

type Cycle struct {
	Name string `json:"name" yaml:"name"`
	UUID string `json:"uuid" yaml:"uuid"`

	Strategy *Strategy `json:"strategy" yaml:"strategy"`

	StartedAt int64 `json:"started_at" yaml:"started_at"`
	DoneAt    int64 `json:"done_at" yaml:"done_at"`
	// The status of the cycle
	Status string `json:"status" yaml:"status"`
}

type Session struct {
	UserID string `json:"user_id" yaml:"user_id"`
}
