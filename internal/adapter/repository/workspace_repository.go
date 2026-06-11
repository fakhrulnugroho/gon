package repository

import (
	"context"
	"gon/internal/adapter/model"
	"gon/internal/core/domain"
	"gon/internal/core/port/driven"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type workspaceRepository struct {
}

func NewWorkspaceRepository() driven.WorkspaceRepository {
	return &workspaceRepository{}
}

func (r *workspaceRepository) Save(ctx context.Context, directory string, workspace domain.Workspace) error {
	data, err := yaml.Marshal(model.NewWorkspaceModelFromDomain(workspace))
	if err != nil {
		return err
	}
	gonDirectory := filepath.Join(directory, ".gon")
	if err := os.MkdirAll(gonDirectory, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(gonDirectory, "workspace.yaml"), data, 0644); err != nil {
		return err
	}
	return nil
}

func (r *workspaceRepository) Load(ctx context.Context, directory string) (*domain.Workspace, error) {
	gonDirectory := filepath.Join(directory, ".gon")
	data, err := os.ReadFile(filepath.Join(gonDirectory, "workspace.yaml"))
	if err != nil {
		return nil, err
	}
	var workspaceModel model.WorkspaceModel
	if err := yaml.Unmarshal(data, &workspaceModel); err != nil {
		return nil, err
	}
	return workspaceModel.ToDomain(), nil
}
