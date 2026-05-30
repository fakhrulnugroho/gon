package repository

import (
	"gon/internal/adapter/model"
	"gon/internal/core/domain"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type workspaceRepository struct {
}

func NewWorkspaceRepository() *workspaceRepository {
	return &workspaceRepository{}
}

func (r *workspaceRepository) Save(directory string, workspace domain.Workspace) error {
	data, err := yaml.Marshal(model.NewWorkspaceModelFromDomain(workspace))
	if err != nil {
		return err
	}
	gonDirectory := filepath.Join(directory, ".gon")
	os.Mkdir(gonDirectory, 0755)
	os.WriteFile(filepath.Join(gonDirectory, "workspace.yaml"), data, 0755)
	return nil
}

func (r *workspaceRepository) Load(directory string) (*domain.Workspace, error) {
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
