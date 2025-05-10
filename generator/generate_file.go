package generator

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/songvi/robo/generator/file"
	"github.com/songvi/robo/models"
)

// GenerateFile creates It creates a new file based on the FileStrategy configuration
func GenerateFile(strategy models.FileStrategy, repositoryPath string) (models.File, error) {
	rand.Seed(time.Now().UnixNano())

	// Validate strategy
	if len(strategy.FileExtension) == 0 || len(strategy.FileExtensionProbability) == 0 ||
		len(strategy.FileSize) == 0 || len(strategy.FileSizeProbability) == 0 ||
		len(strategy.FileLang) == 0 || len(strategy.FileLangNameProbability) == 0 {
		return models.File{}, fmt.Errorf("invalid FileStrategy: one or more required fields are empty")
	}
	if len(strategy.FileExtension) != len(strategy.FileExtensionProbability) ||
		len(strategy.FileSize) != len(strategy.FileSizeProbability) ||
		len(strategy.FileLang) != len(strategy.FileLangNameProbability) {
		return models.File{}, fmt.Errorf("invalid FileStrategy: field lengths do not match their corresponding probabilities")
	}

	// Select file extension based on probability
	extIndex := selectFileIndexByProbability(strategy.FileExtensionProbability)
	fileExtension := strategy.FileExtension[extIndex]

	// Select file size based on probability
	sizeIndex := selectFileIndexByProbability(strategy.FileSizeProbability)
	fileSize := strategy.FileSize[sizeIndex]

	// Select file name language based on probability
	langIndex := selectFileIndexByProbability(strategy.FileLangNameProbability)
	fileLang := strategy.FileLang[langIndex]

	// Generate file name
	fileName := file.GenerateFilename([]string{fileLang})

	// Create file path
	// filePath := filepath.Join("files", fmt.Sprintf("%s.%s", fileName, fileExtension))

	// Create file struct
	generatedFile := models.File{
		Name:          fileName,
		Description:   fmt.Sprintf("Generated %s file in %s", fileExtension, fileLang),
		FileExtension: fileExtension,
		FileSize:      fileSize,
		FileContent:   filepath.Join(repositoryPath, fmt.Sprintf("%s.%s", fileName, fileExtension)),
	}

	// Generate file content
	contentGenerator := file.NewFileContentGenerator(repositoryPath)
	if err := contentGenerator.GenerateContent(&generatedFile, fileLang); err != nil {
		return models.File{}, fmt.Errorf("failed to generate file content: %v", err)
	}

	return generatedFile, nil
}

// selectFileIndexByProbability selects an index based on a probability distribution
func selectFileIndexByProbability(probabilities []float64) int {
	r := rand.Float64()
	sum := 0.0
	for i, p := range probabilities {
		sum += p
		if r <= sum {
			return i
		}
	}
	return len(probabilities) - 1
}
