package service

import (
	"context"
	"testing"

	"gon/internal/core/domain"
	"gon/internal/core/payload"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRequestRepo returns canned request + collections and records Save calls.
type fakeRequestRepo struct {
	request     domain.Request
	collections []domain.Collection
	saved       []string
}

func (f *fakeRequestRepo) Load(ctx context.Context, root, requestPath string) (*domain.Request, []domain.Collection, error) {
	r := f.request
	return &r, f.collections, nil
}
func (f *fakeRequestRepo) Save(ctx context.Context, root, requestPath string, request domain.Request) error {
	f.saved = append(f.saved, requestPath)
	return nil
}
func (f *fakeRequestRepo) Exists(ctx context.Context, root, requestPath string) (bool, error) {
	return false, nil
}

// captureHttpService records the input it was given and returns an empty output.
type captureHttpService struct{ input *payload.HttpExecuteInput }

func (c *captureHttpService) Execute(ctx context.Context, input *payload.HttpExecuteInput) (*payload.HttpExecuteOutput, error) {
	c.input = input
	return &payload.HttpExecuteOutput{StatusCode: 200}, nil
}

func TestRequestServiceRun(t *testing.T) {
	t.Run("merges request, collections, and prefixes paths; overrides win", func(t *testing.T) {
		repo := &fakeRequestRepo{
			request: domain.Request{
				Method:  "POST",
				URL:     "/login",
				Headers: map[string][]string{"Accept": {"application/json"}},
			},
			collections: []domain.Collection{
				{Name: "Admin", Config: domain.Config{Path: "/admin", Headers: map[string]string{"X-Inner": "1"}}},
				{Name: "Auth", Config: domain.Config{Path: "/auth", Headers: map[string]string{"X-Outer": "1", "Accept": "text/plain"}}},
			},
		}
		http := &captureHttpService{}
		svc := NewRequestService(repo, nil, &mockWorkspaceRepository{existsResponse: true}, http)

		overrides := &payload.HttpExecuteInput{Headers: map[string][]string{"X-Inner": {"override"}}}
		_, _, err := svc.Run(context.Background(), "/root", "auth/admin/impersonate", overrides)

		require.NoError(t, err)
		got := http.input
		// path: outermost (/auth) then nearest (/admin) then request url
		assert.Equal(t, "/auth/admin/login", got.URL)
		// override beats collection default
		assert.Equal(t, []string{"override"}, got.Headers["X-Inner"])
		// request header beats collection default for the same key
		assert.Equal(t, []string{"application/json"}, got.Headers["Accept"])
		// outer collection default still applied
		assert.Equal(t, []string{"1"}, got.Headers["X-Outer"])
	})

	t.Run("absolute url bypasses collection path prefixing", func(t *testing.T) {
		repo := &fakeRequestRepo{
			request:     domain.Request{Method: "GET", URL: "https://other.com/x"},
			collections: []domain.Collection{{Config: domain.Config{Path: "/auth"}}},
		}
		http := &captureHttpService{}
		svc := NewRequestService(repo, nil, &mockWorkspaceRepository{existsResponse: true}, http)

		_, _, err := svc.Run(context.Background(), "/root", "auth/x", nil)

		require.NoError(t, err)
		assert.Equal(t, "https://other.com/x", http.input.URL)
	})
}

func TestRequestServiceCreate(t *testing.T) {
	repo := &fakeRequestRepo{}
	collections := &recordingCollectionRepo{}
	svc := NewRequestService(repo, collections, &mockWorkspaceRepository{existsResponse: true}, nil)

	err := svc.Create(context.Background(), "/root", "auth/login", "post")

	require.NoError(t, err)
	// parent collection auth had to be created
	assert.Equal(t, []string{"auth"}, collections.saved)
	assert.Len(t, repo.saved, 1)
}

func TestRequestServiceRequiresWorkspace(t *testing.T) {
	t.Run("Create errors when no workspace is initialized", func(t *testing.T) {
		repo := &fakeRequestRepo{}
		collections := &recordingCollectionRepo{}
		svc := NewRequestService(repo, collections, &mockWorkspaceRepository{existsResponse: false}, nil)

		err := svc.Create(context.Background(), "/root", "auth/login", "post")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no gon workspace found")
		assert.Empty(t, repo.saved)
	})

	t.Run("Run errors when no workspace is initialized", func(t *testing.T) {
		repo := &fakeRequestRepo{request: domain.Request{Method: "GET", URL: "/x"}}
		http := &captureHttpService{}
		svc := NewRequestService(repo, nil, &mockWorkspaceRepository{existsResponse: false}, http)

		_, _, err := svc.Run(context.Background(), "/root", "auth/x", nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no gon workspace found")
		assert.Nil(t, http.input)
	})
}
