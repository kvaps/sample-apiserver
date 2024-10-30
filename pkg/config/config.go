package config

import (
	"fmt"
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

// LoadConfig загружает конфигурацию из указанного пути и выполняет валидацию
func LoadConfig(path string) (*ResourceConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config ResourceConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Валидация конфигурации
	for i, res := range config.Resources {
		if res.Application.Kind == "" {
			return nil, fmt.Errorf("resource at index %d has empty kind", i)
		}
		if res.Application.Plural == "" {
			return nil, fmt.Errorf("resource at index %d has empty plural", i)
		}
		if res.Release.Prefix == "" {
			return nil, fmt.Errorf("resource at index %d has empty release prefix", i)
		}
		if res.Release.Chart.Name == "" {
			return nil, fmt.Errorf("resource at index %d has empty chart name", i)
		}
		if res.Release.Chart.SourceRef.Kind == "" || res.Release.Chart.SourceRef.Name == "" || res.Release.Chart.SourceRef.Namespace == "" {
			return nil, fmt.Errorf("resource at index %d has incomplete chart sourceRef", i)
		}
	}

	return &config, nil
}
