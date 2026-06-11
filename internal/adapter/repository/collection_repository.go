package repository

import (
	"context"
	"os"
	"path/filepath"

	"gon/internal/adapter/model"
	"gon/internal/core/domain"
	"gon/internal/core/port/driven"

	"gopkg.in/yaml.v3"
)

type collectionRepository struct{}

func NewCollectionRepository() driven.CollectionRepository {
	return &collectionRepository{}
}

func (r *collectionRepository) Save(ctx context.Context, root string, collectionPath string, collection domain.Collection) error {
	dir := filepath.Join(root, filepath.Clean(collectionPath))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(model.NewCollectionModelFromDomain(collection))
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, collectionFileName), data, 0644)
}

func (r *collectionRepository) Exists(ctx context.Context, root string, collectionPath string) (bool, error) {
	file := filepath.Join(root, filepath.Clean(collectionPath), collectionFileName)
	if _, err := os.Stat(file); err == nil {
		return true, nil
	}
	return false, nil
}
