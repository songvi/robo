package file

type File struct {
	// The name of the file
	Name string `json:"name" yaml:"name"`
	// The description of the file
	Description string `json:"description" yaml:"description"`
	// The file extension
	FileExtension string `json:"file_extension" yaml:"file_extension"`
	// The file size
	FileSize int `json:"file_size" yaml:"file_size"`
	// The file content
	FileContent string `json:"file_content" yaml:"file_content"`
	// // The file path
	// FilePath string `json:"file_path" yaml:"file_path"`
}
