package repository

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gon/internal/adapter/model"
	"gon/internal/core/domain"
	"gon/internal/core/port/driven"

	"gopkg.in/yaml.v3"
)

const (
	environmentsDirName = "environments"
	activeEnvDirName    = ".gon"
	activeEnvFileName   = "active-env"
)

type environmentRepository struct{}

func NewEnvironmentRepository() driven.EnvironmentRepository {
	return &environmentRepository{}
}

func (r *environmentRepository) envFile(root, name string) string {
	return filepath.Join(root, environmentsDirName, name+".yml")
}

func (r *environmentRepository) Save(ctx context.Context, root string, environment domain.Environment) error {
	dir := filepath.Join(root, environmentsDirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(model.NewEnvironmentModelFromDomain(environment))
	if err != nil {
		return err
	}
	return os.WriteFile(r.envFile(root, environment.Name), data, 0644)
}

func (r *environmentRepository) Load(ctx context.Context, root string, name string) (*domain.Environment, error) {
	data, err := os.ReadFile(r.envFile(root, name))
	if err != nil {
		return nil, err
	}
	var m model.EnvironmentModel
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *environmentRepository) List(ctx context.Context, root string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(root, environmentsDirName))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		base := e.Name()
		if strings.HasSuffix(base, ".yml") {
			names = append(names, strings.TrimSuffix(base, ".yml"))
		} else if strings.HasSuffix(base, ".yaml") {
			names = append(names, strings.TrimSuffix(base, ".yaml"))
		}
	}
	return names, nil
}

func (r *environmentRepository) Exists(ctx context.Context, root string, name string) (bool, error) {
	if _, err := os.Stat(r.envFile(root, name)); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *environmentRepository) ReadActive(ctx context.Context, root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, activeEnvDirName, activeEnvFileName))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (r *environmentRepository) WriteActive(ctx context.Context, root string, name string) error {
	dir := filepath.Join(root, activeEnvDirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, activeEnvFileName), []byte(name+"\n"), 0644)
}
