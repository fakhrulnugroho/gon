package driven

import (
	"context"
	"gon/internal/core/domain"
)

type WorkspaceRepository interface {
	Save(ctx context.Context, directory string, workspace domain.Workspace) error
	Load(ctx context.Context, directory string) (*domain.Workspace, error)
}
