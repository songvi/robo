package models

type Worker struct {
	UUID string `json:"uuid" yaml:"uuid" gorm:"primaryKey;type:uuid;"`
	Name string `json:"name" yaml:"name" gorm:"column:name;type:text;not null"`
}
