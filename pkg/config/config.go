package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// ResourceConfig представляет структуру конфигурационного файла
type ResourceConfig struct {
	Resources []Resource `yaml:"resources"`
}

// Resource описывает отдельный ресурс
type Resource struct {
	Application ApplicationConfig `yaml:"application"`
	Release     ReleaseConfig     `yaml:"release"`
}

// ApplicationConfig содержит настройки приложения
type ApplicationConfig struct {
	Kind       string   `yaml:"kind"`
	Singular   string   `yaml:"singular"`
	Plural     string   `yaml:"plural"`
	ShortNames []string `yaml:"shortNames"`
}

// ReleaseConfig содержит настройки релиза
type ReleaseConfig struct {
	Prefix string            `yaml:"prefix"`
	Labels map[string]string `yaml:"labels"`
	Chart  ChartConfig       `yaml:"chart"`
}

// ChartConfig содержит настройки чарта
type ChartConfig struct {
	Name      string          `yaml:"name"`
	SourceRef SourceRefConfig `yaml:"sourceRef"`
}

// SourceRefConfig содержит ссылку на источник чарта
type SourceRefConfig struct {
	Kind      string `yaml:"kind"`
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

// LoadConfig загружает конфигурацию из указанного пути
func LoadConfig(path string) (*ResourceConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config ResourceConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
