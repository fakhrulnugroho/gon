package service

import (
	"context"
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

func getFolderName(directory string) string {
	if directory == "" {
		return ""
	}
	return strcase.ToKebab(filepath.Base(filepath.Clean(directory)))
}
