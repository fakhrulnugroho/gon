package service

import (
	"context"
	"fmt"
	"gon/internal/core/domain"
	"gon/internal/core/port/driven"
	"gon/internal/core/port/driving"
	"path/filepath"

	"github.com/iancoleman/strcase"
)

type workspaceService struct {
	workspaceRepository driven.WorkspaceRepository
}

func NewWorkspaceService(repo driven.WorkspaceRepository) driving.WorkspaceService {
	return &workspaceService{workspaceRepository: repo}
}

func (s *workspaceService) Create(ctx context.Context, directory string) error {
	workspace := domain.Workspace{
		Name:    getFolderName(directory),
		Config:  domain.Config{},
		BaseURL: "https://api.example.com",
	}
	return s.workspaceRepository.Save(ctx, directory, workspace)
}

// ensureWorkspace guards collection/request operations: they only make sense
// inside an initialized workspace, i.e. a directory with a workspace.yml at its
// root. When no workspace is present it returns an actionable error pointing at
// 'init'.
func ensureWorkspace(ctx context.Context, repo driven.WorkspaceRepository, root string) error {
	exists, err := repo.Exists(ctx, root)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("no gon workspace found, run 'init' first")
	}
	return nil
}

func getFolderName(directory string) string {
	if directory == "" {
		return ""
	}
	return strcase.ToKebab(filepath.Base(filepath.Clean(directory)))
}
