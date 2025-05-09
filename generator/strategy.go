package generator

type UserStrategy struct {
	UserLang        []string  `json:"user_lang" yaml:"user_lang"`
	LangProbability []float64 `json:"lang_probability" yaml:"lang_probability"`
}

type FileStrategy struct {
	FileExtension            []string  `json:"file_extension" yaml:"file_extension"`
	FileExtensionProbability []float64 `json:"file_extension_probability" yaml:"file_extension_probability"`
	FileSize                 []int     `json:"file_size" yaml:"file_size"`
	FileSizeProbability      []float64 `json:"file_size_probability" yaml:"file_size_probability"`
	FileLang                 []string  `json:"file_name_lang" yaml:"file_name_lang"`
	FileLangNameProbability  []float64 `json:"file_name_probability" yaml:"file_name_probability"`
}

type WorkspaceStrategy struct {
	NumberOfUsers            []int     `json:"number_of_users" yaml:"number_of_users"`
	NumberOfUsersProbability []float64 `json:"number_of_users_probability" yaml:"number_of_users_probability"`
}
