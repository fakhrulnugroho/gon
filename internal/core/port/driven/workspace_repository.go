package driven

import "gon/internal/core/domain"

type WorkspaceRepository interface {
	Save(directory string, workspace domain.Workspace) error
	Load(directory string) (*domain.Workspace, error)
}
