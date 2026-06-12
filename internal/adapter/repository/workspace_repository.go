package repository

import (
	"context"
	"errors"
	"gon/internal/adapter/model"
	"gon/internal/core/domain"
	"gon/internal/core/port/driven"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// gonDir is the per-project directory that holds the workspace, collections,
// requests, and cache. Every gon artifact lives under it.
const gonDir = ".gon"

const workspaceFileName = "workspace.yaml"

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
	gonDirectory := filepath.Join(directory, gonDir)
	if err := os.MkdirAll(gonDirectory, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(gonDirectory, workspaceFileName), data, 0644); err != nil {
		return err
	}
	return nil
}

func (r *workspaceRepository) Load(ctx context.Context, directory string) (*domain.Workspace, error) {
	gonDirectory := filepath.Join(directory, gonDir)
	data, err := os.ReadFile(filepath.Join(gonDirectory, workspaceFileName))
	if err != nil {
		return nil, err
	}
	var workspaceModel model.WorkspaceModel
	if err := yaml.Unmarshal(data, &workspaceModel); err != nil {
		return nil, err
	}
	return workspaceModel.ToDomain(), nil
}

// Exists reports whether a workspace has been initialized for directory, i.e.
// whether directory/.gon/workspace.yaml is present.
func (r *workspaceRepository) Exists(ctx context.Context, directory string) (bool, error) {
	if _, err := os.Stat(filepath.Join(directory, gonDir, workspaceFileName)); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
