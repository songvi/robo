package workspace

type Workspace struct {
	// The name of the workspace
	Name  string   `json:"name" yaml:"name"`
	Users []string `json:"users" yaml:"users"`
}
