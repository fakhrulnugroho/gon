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

// workspaceFileName is the manifest that marks a directory as a gon workspace.
// It lives at the workspace root, alongside collections and requests, so the
// whole folder is a self-contained, shareable artifact.
const workspaceFileName = "workspace.yml"

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
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(directory, workspaceFileName), data, 0644); err != nil {
		return err
	}
	return nil
}

func (r *workspaceRepository) Load(ctx context.Context, directory string) (*domain.Workspace, error) {
	data, err := os.ReadFile(filepath.Join(directory, workspaceFileName))
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
// whether directory/workspace.yml is present.
func (r *workspaceRepository) Exists(ctx context.Context, directory string) (bool, error) {
	if _, err := os.Stat(filepath.Join(directory, workspaceFileName)); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
