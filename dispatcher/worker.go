package dispatcher

type Worker struct {
	// The name of the worker
	Name string `json:"name" yaml:"name"`
	UUID string `json:"uuid" yaml:"uuid"`
}
