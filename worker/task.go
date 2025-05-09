package worker

type Task struct {
	// The name of the task
	Name string `json:"name" yaml:"name"`
	UUID string `json:"uuid" yaml:"uuid"`
}

type TaskResult struct {
	UUID string `json:"uuid" yaml:"uuid"`
	Task *Task  `json:"task" yaml:"task"`
	// The name of the task
	Name string `json:"name" yaml:"name"`
	// The result of the task
	Result string `json:"result" yaml:"result"`
	// The error of the task
	Error string `json:"error" yaml:"error"`
	// The start time of the task
	StartAt int64 `json:"start_at" yaml:"start_at"`
	// The end time of the task
	EndAt int64 `json:"end_at" yaml:"end_at"`
	// The status of the task
	Status string `json:"status" yaml:"status"`
	// The input data of the task
	InputData string `json:"input_data" yaml:"input_data"`
	// The output data of the task
	OutputData string `json:"output_data" yaml:"output_data"`
}
