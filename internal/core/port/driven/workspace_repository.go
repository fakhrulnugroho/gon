package driven

import "gon/internal/core/domain"

type WorkspaceRepository interface {
	Save(workspace domain.Workspace) error
}
