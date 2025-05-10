package useraction

type UserAction struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	// The action type
	ActionType string `json:"action_type" yaml:"action_type"`

	CycleID   string `json:"cycle_id" yaml:"cycle_id"`
	SessionID string `json:"session_id" yaml:"session_id"`
}
