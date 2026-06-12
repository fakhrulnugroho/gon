package driven

import (
	"context"
	"gon/internal/core/domain"
)

type WorkspaceRepository interface {
	Save(ctx context.Context, directory string, workspace domain.Workspace) error
	Load(ctx context.Context, directory string) (*domain.Workspace, error)
	Exists(ctx context.Context, directory string) (bool, error)
	// EnsureGitignore makes sure each entry is present in directory/.gitignore,
	// creating the file if needed and never duplicating an existing line.
	EnsureGitignore(ctx context.Context, directory string, entries []string) error
}
