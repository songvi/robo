package models

type UserStrategy struct {
	UserLang        []string  `json:"user_lang" yaml:"user_lang"`
	LangProbability []float64 `json:"lang_probability" yaml:"lang_probability"`
}
