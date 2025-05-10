package models

import "encoding/json"

type Job struct {
	UUID     string `json:"uuid" yaml:"uuid"`
	WorkerID string `json:"worker_id" yaml:"worker_id"`
	// The name of the job
	Name       string          `json:"name" yaml:"name"`
	InputData  json.RawMessage `json:"input_data" yaml:"input_data"`
	OutputData json.RawMessage `json:"output_data" yaml:"output_data"`

	Error   string `json:"error" yaml:"error"`
	StartAt int64  `json:"start_at" yaml:"start_at"`
	DoneAt  int64  `json:"done_at" yaml:"done_at"`
	// The status of the job
	Status string `json:"status" yaml:"status"`

	CycleUUID string `json:"cycle_uuid" yaml:"cycle_uuid"`
	SessionID string `json:"session_id" yaml:"session_id"`
}
