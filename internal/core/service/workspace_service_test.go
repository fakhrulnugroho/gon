package service

import (
	"context"
	"errors"
	"testing"

	"gon/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockWorkspaceRepository is a hand-written test double implementing
// driven.WorkspaceRepository. It records the arguments passed to Save and
// returns a configurable error.
type mockWorkspaceRepository struct {
	saveErr        error
	savedDir       string
	saved          domain.Workspace
	saveCalls      int
	loadErr        error
	loadResponse   *domain.Workspace
	existsResponse bool
	existsErr      error
}

func (m *mockWorkspaceRepository) Save(_ context.Context, directory string, workspace domain.Workspace) error {
	m.saveCalls++
	m.savedDir = directory
	m.saved = workspace
	return m.saveErr
}

func (m *mockWorkspaceRepository) Load(_ context.Context, _ string) (*domain.Workspace, error) {
	return m.loadResponse, m.loadErr
}

func (m *mockWorkspaceRepository) Exists(_ context.Context, _ string) (bool, error) {
	return m.existsResponse, m.existsErr
}

func TestWorkspaceServiceCreate(t *testing.T) {
	repo := &mockWorkspaceRepository{}
	envRepo := newFakeEnvRepo()
	svc := NewWorkspaceService(repo, envRepo)

	err := svc.Create(context.Background(), "/home/user/My Project")
	require.NoError(t, err)

	assert.Equal(t, 1, repo.saveCalls)
	assert.Equal(t, "/home/user/My Project", repo.savedDir)
	assert.Equal(t, "my-project", repo.saved.Name)
	// base_url is no longer written to the workspace; it lives in the environment.
	assert.Equal(t, "", repo.saved.BaseURL)
	assert.Equal(t, domain.Config{}, repo.saved.Config)

	// a 'local' environment is scaffolded and marked active.
	local, ok := envRepo.envs["local"]
	require.True(t, ok)
	assert.Equal(t, "https://api.example.com", local.BaseURL)
	assert.Equal(t, "local", envRepo.active)
}

func TestWorkspaceServiceCreatePropagatesError(t *testing.T) {
	sentinel := errors.New("disk full")
	repo := &mockWorkspaceRepository{saveErr: sentinel}
	svc := NewWorkspaceService(repo, newFakeEnvRepo())

	err := svc.Create(context.Background(), "/tmp/proj")
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
}

func TestGetFolderName(t *testing.T) {
	tests := []struct {
		name      string
		directory string
		want      string
	}{
		{"empty returns empty", "", ""},
		{"simple folder", "/home/user/project", "project"},
		{"converts to kebab case", "/home/user/MyCoolApp", "my-cool-app"},
		{"spaces become dashes", "/home/user/My Project", "my-project"},
		{"trailing slash ignored", "/home/user/api/", "api"},
		{"single segment", "project", "project"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, getFolderName(tt.directory))
		})
	}
}
