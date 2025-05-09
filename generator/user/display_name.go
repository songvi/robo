package user

import (
	"github.com/songvi/robo/generator/file"
)

func GenerateDisplayName(strategy UserStrategy) string {
	return file.GenerateFilename(strategy.UserLang)
}
