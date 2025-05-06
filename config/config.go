package config

import (
	_ "embed"
	"fmt"

	yaml "gopkg.in/yaml.v3"
)

var (
	//go:embed development/config.yml
	developmentConfig []byte
	//go:embed staging/config.yml
	stagingConfig []byte
	//go:embed production/config.yml
	productionConfig []byte
	//go:embed testing/config.yml
	testConfig []byte
)

type (
	Config struct {
		Env   string  `yaml:"env"`
		Port  int     `yaml:"port"`
		Store Storage `yaml:"store"`
	}
	Storage struct {
		Kind      string
		AssetPath string `yaml:"assetPath"`
		TmpPath   string `yaml:"tmpPath"`
	}
)

func Load(env string) (*Config, error) {
	conf := &Config{}
	var data []byte
	switch env {
	case "development", "dev":
		data = developmentConfig
	case "staging", "stag":
		data = stagingConfig
	case "production", "prod":
		data = productionConfig
	case "testing", "test":
		data = testConfig
	default:
		return nil, fmt.Errorf("unknown config environment %v", env)
	}
	return conf, yaml.Unmarshal(data, conf)
}
