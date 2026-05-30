package model

import "gon/internal/core/domain"

type WorkspaceModel struct {
	Name    string      `yaml:"name,omitempty"`
	BaseURL string      `yaml:"base_url,omitempty"`
	Config  ConfigModel `yaml:"config,omitempty"`
}

func NewWorkspaceModelFromDomain(workspace domain.Workspace) *WorkspaceModel {
	return &WorkspaceModel{
		Name:    workspace.Name,
		BaseURL: workspace.BaseURL,
		Config:  *NewConfigModelFromDomain(workspace.Config),
	}
}

func (m *WorkspaceModel) ToDomain() *domain.Workspace {
	return &domain.Workspace{
		Name:    m.Name,
		BaseURL: m.BaseURL,
		Config:  m.Config.ToDomain(),
	}
}
