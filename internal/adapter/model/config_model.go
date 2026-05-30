package model

import "gon/internal/core/domain"

type ConfigModel struct {
	Path    string            `yaml:"path,omitempty"`
	Query   map[string]string `yaml:"query,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

func NewConfigModelFromDomain(config domain.Config) *ConfigModel {
	return &ConfigModel{
		Path:    config.Path,
		Query:   config.Query,
		Headers: config.Headers,
	}
}

func (m *ConfigModel) ToDomain() domain.Config {
	return domain.Config{
		Path:    m.Path,
		Query:   m.Query,
		Headers: m.Headers,
	}
}
