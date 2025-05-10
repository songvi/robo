package user

import (
	"github.com/songvi/robo/generator/file"
	"github.com/songvi/robo/models"
)

func GenerateDisplayName(strategy models.UserStrategy) string {
	return file.GenerateFilename(strategy.UserLang)
}
