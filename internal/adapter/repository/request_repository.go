package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gon/internal/adapter/model"
	"gon/internal/core/domain"
	"gon/internal/core/port/driven"

	"gopkg.in/yaml.v3"
)

const collectionFileName = "collection.yml"

type requestRepository struct{}

func NewRequestRepository() driven.RequestRepository {
	return &requestRepository{}
}

// resolveFile returns the on-disk path of a request file, trying .yml then
// .yaml. It returns ok=false when neither exists.
func resolveFile(root, requestPath string) (string, bool) {
	clean := filepath.Clean(requestPath)
	for _, ext := range []string{".yml", ".yaml"} {
		candidate := filepath.Join(root, clean) + ext
		if hasExtension(clean) {
			candidate = filepath.Join(root, clean)
		}
		if _, err := os.Stat(candidate); err == nil {
			return candidate, true
		}
		if hasExtension(clean) {
			break
		}
	}
	return "", false
}

func hasExtension(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".yml" || ext == ".yaml"
}

func (r *requestRepository) Load(ctx context.Context, root string, requestPath string) (*domain.Request, []domain.Collection, error) {
	clean := filepath.Clean(requestPath)
	base := strings.TrimSuffix(filepath.Base(clean), filepath.Ext(clean))
	if base == "collection" {
		return nil, nil, fmt.Errorf("%q is a reserved collection file, not a request", requestPath)
	}

	file, ok := resolveFile(root, requestPath)
	if !ok {
		return nil, nil, fmt.Errorf("request not found: %s", requestPath)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, nil, err
	}
	var requestModel model.RequestModel
	if err := yaml.Unmarshal(data, &requestModel); err != nil {
		return nil, nil, fmt.Errorf("error parsing %s: %w", file, err)
	}
	request, err := requestModel.ToDomain()
	if err != nil {
		return nil, nil, fmt.Errorf("error in %s: %w", file, err)
	}

	collections, err := loadCollectionChain(root, filepath.Dir(file))
	if err != nil {
		return nil, nil, err
	}
	return request, collections, nil
}

// loadCollectionChain walks from dir up to and including root, collecting any
// collection.yml found, nearest-first.
func loadCollectionChain(root, dir string) ([]domain.Collection, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	current, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	var collections []domain.Collection
	for {
		file := filepath.Join(current, collectionFileName)
		if data, err := os.ReadFile(file); err == nil {
			var collectionModel model.CollectionModel
			if err := yaml.Unmarshal(data, &collectionModel); err != nil {
				return nil, fmt.Errorf("error parsing %s: %w", file, err)
			}
			collections = append(collections, *collectionModel.ToDomain())
		}
		if current == absRoot {
			break
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return collections, nil
}

func (r *requestRepository) Save(ctx context.Context, root string, requestPath string, request domain.Request) error {
	clean := filepath.Clean(requestPath)
	if !hasExtension(clean) {
		clean += ".yml"
	}
	target := filepath.Join(root, clean)
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(model.NewRequestModelFromDomain(request))
	if err != nil {
		return err
	}
	return os.WriteFile(target, data, 0644)
}

func (r *requestRepository) Exists(ctx context.Context, root string, requestPath string) (bool, error) {
	_, ok := resolveFile(root, requestPath)
	return ok, nil
}
