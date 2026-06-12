package service

import (
	"context"
	"fmt"
	"path"
	"strings"

	"gon/internal/core/domain"
	"gon/internal/core/port/driven"
	"gon/internal/core/port/driving"

	"github.com/iancoleman/strcase"
)

type collectionService struct {
	collectionRepository driven.CollectionRepository
	workspaceRepository  driven.WorkspaceRepository
}

func NewCollectionService(collectionRepository driven.CollectionRepository, workspaceRepository driven.WorkspaceRepository) driving.CollectionService {
	return &collectionService{collectionRepository: collectionRepository, workspaceRepository: workspaceRepository}
}

func (s *collectionService) Create(ctx context.Context, root string, collectionPath string) error {
	if err := ensureWorkspace(ctx, s.workspaceRepository, root); err != nil {
		return err
	}

	normalized := strings.Trim(toSlash(collectionPath), "/")
	if normalized == "" {
		return fmt.Errorf("collection path is required")
	}

	segments := strings.Split(normalized, "/")
	for i := range segments {
		sub := strings.Join(segments[:i+1], "/")
		isTarget := i == len(segments)-1

		exists, err := s.collectionRepository.Exists(ctx, root, sub)
		if err != nil {
			return err
		}
		if exists {
			if isTarget {
				return fmt.Errorf("collection already exists: %s", sub)
			}
			continue
		}

		name := strcase.ToKebab(path.Base(sub))
		if err := s.collectionRepository.Save(ctx, root, sub, domain.Collection{Name: name}); err != nil {
			return err
		}
	}
	return nil
}
