package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Project             string   `yaml:"project"`
	ExtraExcludePattern []string `yaml:"extraExcludePattern"`
	ExtraExcludeFileExt []string `yaml:"extraExcludeFileExt"`
	Source              string   `yaml:"source"`
	Target              string   `yaml:"target"`
	fileNameTemplate    *string  `yaml:"fileNameTemplate,omitempty"`
}

func NewConfig(filepath string) *Config {
	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	config := &Config{}
	err = yaml.NewDecoder(f).Decode(config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}
