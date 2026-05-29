package repository

import (
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

func (r *workspaceRepository) Save(workspace domain.Workspace) error {
	data, err := yaml.Marshal(workspace)
	if err != nil {
		return err
	}
	gonDirectory := filepath.Join(workspace.WorkingDirectory, ".gon")
	os.Mkdir(gonDirectory, 0755)
	os.WriteFile(filepath.Join(gonDirectory, "workspace.yaml"), data, 0755)
	return nil
}
