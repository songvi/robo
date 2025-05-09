package workspace

import "github.com/songvi/robo/generator/user"

type Workspace struct {
	// The name of the workspace
	Name  string       `json:"name" yaml:"name"`
	Users []*user.User `json:"users" yaml:"users"`
}
