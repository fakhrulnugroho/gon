package model

import "gon/internal/core/domain"

type EnvironmentModel struct {
	Name      string            `yaml:"name,omitempty"`
	BaseURL   string            `yaml:"base_url,omitempty"`
	Variables map[string]string `yaml:"variables,omitempty"`
}

func NewEnvironmentModelFromDomain(environment domain.Environment) *EnvironmentModel {
	return &EnvironmentModel{
		Name:      environment.Name,
		BaseURL:   environment.BaseURL,
		Variables: environment.Variables,
	}
}

func (m *EnvironmentModel) ToDomain() *domain.Environment {
	return &domain.Environment{
		Name:      m.Name,
		BaseURL:   m.BaseURL,
		Variables: m.Variables,
	}
}
