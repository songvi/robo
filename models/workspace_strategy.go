package models

type WorkspaceStrategy struct {
	NumberOfUsers            []int     `json:"number_of_users" yaml:"number_of_users"`
	NumberOfUsersProbability []float64 `json:"number_of_users_probability" yaml:"number_of_users_probability"`
}
