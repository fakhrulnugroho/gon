package service

import (
	"context"
	"fmt"
	"path"
	"strings"

	"gon/internal/core/domain"
	"gon/internal/core/payload"
	"gon/internal/core/port/driven"
	"gon/internal/core/port/driving"

	"github.com/iancoleman/strcase"
)

type requestService struct {
	requestRepository    driven.RequestRepository
	collectionRepository driven.CollectionRepository
	httpService          driving.HttpService
}

func NewRequestService(
	requestRepository driven.RequestRepository,
	collectionRepository driven.CollectionRepository,
	httpService driving.HttpService,
) driving.RequestService {
	return &requestService{
		requestRepository:    requestRepository,
		collectionRepository: collectionRepository,
		httpService:          httpService,
	}
}

func (s *requestService) Run(ctx context.Context, root string, requestPath string, overrides *payload.HttpExecuteInput) (*payload.HttpExecuteOutput, error) {
	request, collections, err := s.requestRepository.Load(ctx, root, requestPath)
	if err != nil {
		return nil, err
	}

	input, err := request.ToInput()
	if err != nil {
		return nil, err
	}

	applyOverrides(input, overrides)

	// Collection defaults: nearest-first so inner collections win (additive).
	for i := range collections {
		collections[i].Config.ApplyDefaults(input)
	}

	input.URL = prefixCollectionPaths(input.URL, collections)

	return s.httpService.Execute(ctx, input)
}

// applyOverrides copies per-execution values onto input, replacing existing
// keys so the override always wins.
func applyOverrides(input *payload.HttpExecuteInput, overrides *payload.HttpExecuteInput) {
	if overrides == nil {
		return
	}
	if input.Headers == nil {
		input.Headers = make(map[string][]string)
	}
	for key, values := range overrides.Headers {
		input.Headers[key] = values
	}
	if input.Query == nil {
		input.Query = make(map[string][]string)
	}
	for key, values := range overrides.Query {
		input.Query[key] = values
	}
	if overrides.Body != nil {
		input.Body = overrides.Body
	}
}

// prefixCollectionPaths prepends each collection's configured path, outermost
// first, to a relative URL. Absolute URLs are returned unchanged.
func prefixCollectionPaths(url string, collections []domain.Collection) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	var prefix string
	// collections is nearest-first; iterate in reverse for outermost-first.
	for i := len(collections) - 1; i >= 0; i-- {
		prefix += collections[i].Config.Path
	}
	return prefix + url
}

func (s *requestService) Create(ctx context.Context, root string, requestPath string, method string) error {
	exists, err := s.requestRepository.Exists(ctx, root, requestPath)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("request already exists: %s", requestPath)
	}

	// Ensure the parent folder is a collection.
	parent := path.Dir(filepath_ToSlash(requestPath))
	if parent != "." && parent != "" {
		ok, err := s.collectionRepository.Exists(ctx, root, parent)
		if err != nil {
			return err
		}
		if !ok {
			name := strcase.ToKebab(path.Base(parent))
			if err := s.collectionRepository.Save(ctx, root, parent, domain.Collection{Name: name}); err != nil {
				return err
			}
		}
	}

	name := strcase.ToKebab(strings.TrimSuffix(path.Base(filepath_ToSlash(requestPath)), ".yml"))
	request := domain.Request{
		Name:   name,
		Method: strings.ToUpper(method),
		URL:    "/",
	}
	return s.requestRepository.Save(ctx, root, requestPath, request)
}

// filepath_ToSlash normalizes OS separators to forward slashes so path.Dir /
// path.Base behave consistently regardless of platform.
func filepath_ToSlash(p string) string {
	return strings.ReplaceAll(p, "\\", "/")
}
