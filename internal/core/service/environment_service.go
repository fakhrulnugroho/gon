package service

import (
	"context"
	"fmt"

	"gon/internal/core/domain"
	"gon/internal/core/port/driven"
	"gon/internal/core/port/driving"

	"github.com/iancoleman/strcase"
)

type environmentService struct {
	environmentRepository driven.EnvironmentRepository
	workspaceRepository   driven.WorkspaceRepository
}

func NewEnvironmentService(environmentRepository driven.EnvironmentRepository, workspaceRepository driven.WorkspaceRepository) driving.EnvironmentService {
	return &environmentService{
		environmentRepository: environmentRepository,
		workspaceRepository:   workspaceRepository,
	}
}

func (s *environmentService) Create(ctx context.Context, root string, name string) error {
	if err := ensureWorkspace(ctx, s.workspaceRepository, root); err != nil {
		return err
	}
	normalized := strcase.ToKebab(name)
	if normalized == "" {
		return fmt.Errorf("environment name is required")
	}
	exists, err := s.environmentRepository.Exists(ctx, root, normalized)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("environment already exists: %s", normalized)
	}
	env := domain.Environment{
		Name:      normalized,
		BaseURL:   "https://api.example.com",
		Variables: map[string]string{},
	}
	return s.environmentRepository.Save(ctx, root, env)
}

func (s *environmentService) List(ctx context.Context, root string) ([]string, string, error) {
	names, err := s.environmentRepository.List(ctx, root)
	if err != nil {
		return nil, "", err
	}
	active, err := s.environmentRepository.ReadActive(ctx, root)
	if err != nil {
		return nil, "", err
	}
	return names, active, nil
}

func (s *environmentService) Use(ctx context.Context, root string, name string) error {
	if err := ensureWorkspace(ctx, s.workspaceRepository, root); err != nil {
		return err
	}
	exists, err := s.environmentRepository.Exists(ctx, root, name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("environment not found: %s", name)
	}
	return s.environmentRepository.WriteActive(ctx, root, name)
}

func (s *environmentService) Resolve(ctx context.Context, root string, override string) (*domain.Environment, error) {
	if override != "" {
		return s.environmentRepository.Load(ctx, root, override)
	}
	active, err := s.environmentRepository.ReadActive(ctx, root)
	if err != nil {
		return nil, err
	}
	if active != "" {
		return s.environmentRepository.Load(ctx, root, active)
	}
	names, err := s.environmentRepository.List(ctx, root)
	if err != nil {
		return nil, err
	}
	switch len(names) {
	case 0:
		return nil, nil
	case 1:
		return s.environmentRepository.Load(ctx, root, names[0])
	default:
		return nil, fmt.Errorf("no active environment; run 'env use <name>' or pass --env")
	}
}
