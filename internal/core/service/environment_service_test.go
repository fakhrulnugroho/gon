package service

import (
	"context"
	"testing"

	"gon/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeEnvRepo is an in-memory EnvironmentRepository test double.
type fakeEnvRepo struct {
	envs   map[string]domain.Environment
	active string
}

func newFakeEnvRepo() *fakeEnvRepo {
	return &fakeEnvRepo{envs: map[string]domain.Environment{}}
}

func (f *fakeEnvRepo) Save(_ context.Context, _ string, environment domain.Environment) error {
	f.envs[environment.Name] = environment
	return nil
}
func (f *fakeEnvRepo) Load(_ context.Context, _ string, name string) (*domain.Environment, error) {
	env, ok := f.envs[name]
	if !ok {
		return nil, assertMissing(name)
	}
	e := env
	return &e, nil
}
func (f *fakeEnvRepo) List(_ context.Context, _ string) ([]string, error) {
	names := make([]string, 0, len(f.envs))
	for n := range f.envs {
		names = append(names, n)
	}
	return names, nil
}
func (f *fakeEnvRepo) Exists(_ context.Context, _ string, name string) (bool, error) {
	_, ok := f.envs[name]
	return ok, nil
}
func (f *fakeEnvRepo) ReadActive(_ context.Context, _ string) (string, error) { return f.active, nil }
func (f *fakeEnvRepo) WriteActive(_ context.Context, _ string, name string) error {
	f.active = name
	return nil
}

func assertMissing(name string) error { return &missingEnvErr{name} }

type missingEnvErr struct{ name string }

func (e *missingEnvErr) Error() string { return "environment not found: " + e.name }

func TestEnvironmentServiceResolve(t *testing.T) {
	ctx := context.Background()

	t.Run("override flag wins", func(t *testing.T) {
		repo := newFakeEnvRepo()
		repo.envs["dev"] = domain.Environment{Name: "dev", BaseURL: "https://dev"}
		repo.envs["prod"] = domain.Environment{Name: "prod", BaseURL: "https://prod"}
		repo.active = "dev"
		svc := NewEnvironmentService(repo, &mockWorkspaceRepository{existsResponse: true})

		got, err := svc.Resolve(ctx, "/root", "prod")
		require.NoError(t, err)
		assert.Equal(t, "prod", got.Name)
	})

	t.Run("falls back to persisted active", func(t *testing.T) {
		repo := newFakeEnvRepo()
		repo.envs["dev"] = domain.Environment{Name: "dev"}
		repo.envs["prod"] = domain.Environment{Name: "prod"}
		repo.active = "dev"
		svc := NewEnvironmentService(repo, &mockWorkspaceRepository{existsResponse: true})

		got, err := svc.Resolve(ctx, "/root", "")
		require.NoError(t, err)
		assert.Equal(t, "dev", got.Name)
	})

	t.Run("single environment is used when none active", func(t *testing.T) {
		repo := newFakeEnvRepo()
		repo.envs["local"] = domain.Environment{Name: "local"}
		svc := NewEnvironmentService(repo, &mockWorkspaceRepository{existsResponse: true})

		got, err := svc.Resolve(ctx, "/root", "")
		require.NoError(t, err)
		assert.Equal(t, "local", got.Name)
	})

	t.Run("zero environments returns nil", func(t *testing.T) {
		repo := newFakeEnvRepo()
		svc := NewEnvironmentService(repo, &mockWorkspaceRepository{existsResponse: true})

		got, err := svc.Resolve(ctx, "/root", "")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("multiple environments and none active errors", func(t *testing.T) {
		repo := newFakeEnvRepo()
		repo.envs["dev"] = domain.Environment{Name: "dev"}
		repo.envs["prod"] = domain.Environment{Name: "prod"}
		svc := NewEnvironmentService(repo, &mockWorkspaceRepository{existsResponse: true})

		_, err := svc.Resolve(ctx, "/root", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no active environment")
	})
}

func TestEnvironmentServiceCreateRequiresWorkspace(t *testing.T) {
	repo := newFakeEnvRepo()
	svc := NewEnvironmentService(repo, &mockWorkspaceRepository{existsResponse: false})

	err := svc.Create(context.Background(), "/root", "dev")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no gon workspace found")
}

func TestEnvironmentServiceCreateAndUse(t *testing.T) {
	ctx := context.Background()
	repo := newFakeEnvRepo()
	svc := NewEnvironmentService(repo, &mockWorkspaceRepository{existsResponse: true})

	require.NoError(t, svc.Create(ctx, "/root", "Dev"))
	_, ok := repo.envs["dev"]
	assert.True(t, ok, "name should be kebab-cased")

	require.NoError(t, svc.Use(ctx, "/root", "dev"))
	assert.Equal(t, "dev", repo.active)

	require.Error(t, svc.Use(ctx, "/root", "missing"))
}

func TestEnvironmentServiceList(t *testing.T) {
	ctx := context.Background()
	repo := newFakeEnvRepo()
	repo.envs["dev"] = domain.Environment{Name: "dev"}
	repo.active = "dev"
	svc := NewEnvironmentService(repo, &mockWorkspaceRepository{existsResponse: true})

	names, active, err := svc.List(ctx, "/root")
	require.NoError(t, err)
	assert.Equal(t, []string{"dev"}, names)
	assert.Equal(t, "dev", active)
}
