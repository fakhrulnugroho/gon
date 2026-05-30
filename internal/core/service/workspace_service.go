package service

import (
	"gon/internal/core/domain"
	"gon/internal/core/port/driven"
	"gon/internal/core/port/driving"

	"github.com/iancoleman/strcase"
)

type workspaceService struct {
	workspaceRepository driven.WorkspaceRepository
}

func NewWorkspaceService(repo driven.WorkspaceRepository) driving.WorkspaceService {
	return &workspaceService{workspaceRepository: repo}
}

func (s *workspaceService) Create(directory string) error {
	workspace := domain.Workspace{
		Name:    getFolderName(directory),
		Config:  domain.Config{},
		BaseURL: "https://api.example.com",
	}
	return s.workspaceRepository.Save(directory, workspace)
}

func getFolderName(directory string) string {
	if directory == "" {
		return ""
	}

	// hapus trailing '/'
	for len(directory) > 1 && directory[len(directory)-1] == '/' {
		directory = directory[:len(directory)-1]
	}

	lastSlash := -1
	for i := len(directory) - 1; i >= 0; i-- {
		if directory[i] == '/' {
			lastSlash = i
			break
		}
	}

	if lastSlash == -1 {
		return strcase.ToKebab(directory)
	}

	return strcase.ToKebab(directory[lastSlash+1:])
}
