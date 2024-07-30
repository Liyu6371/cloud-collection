package config

type Logger struct {
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
}
