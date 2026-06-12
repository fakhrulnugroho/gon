package service

import (
	"context"
	"fmt"
	"net/textproto"
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
	workspaceRepository  driven.WorkspaceRepository
	httpService          driving.HttpService
}

func NewRequestService(
	requestRepository driven.RequestRepository,
	collectionRepository driven.CollectionRepository,
	workspaceRepository driven.WorkspaceRepository,
	httpService driving.HttpService,
) driving.RequestService {
	return &requestService{
		requestRepository:    requestRepository,
		collectionRepository: collectionRepository,
		workspaceRepository:  workspaceRepository,
		httpService:          httpService,
	}
}

func (s *requestService) Run(ctx context.Context, root string, requestPath string, overrides *payload.HttpExecuteInput) (*payload.HttpExecuteInput, *payload.HttpExecuteOutput, error) {
	if err := ensureWorkspace(ctx, s.workspaceRepository, root); err != nil {
		return nil, nil, err
	}

	request, collections, err := s.requestRepository.Load(ctx, root, requestPath)
	if err != nil {
		return nil, nil, err
	}

	input, err := request.ToInput()
	if err != nil {
		return nil, nil, err
	}

	applyOverrides(input, overrides)

	// Collection defaults: nearest-first so inner collections win (additive).
	for i := range collections {
		collections[i].Config.ApplyDefaults(input)
	}

	input.URL = prefixCollectionPaths(input.URL, collections)

	result, err := s.httpService.Execute(ctx, input, nil)
	return input, result, err
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
		input.Headers[textproto.CanonicalMIMEHeaderKey(key)] = values
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
	if err := ensureWorkspace(ctx, s.workspaceRepository, root); err != nil {
		return err
	}

	exists, err := s.requestRepository.Exists(ctx, root, requestPath)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("request already exists: %s", requestPath)
	}

	// Ensure the parent folder is a collection.
	parent := path.Dir(toSlash(requestPath))
	if parent != "." && parent != "" {
		if err := s.ensureCollections(ctx, root, parent); err != nil {
			return err
		}
	}

	base := path.Base(toSlash(requestPath))
	base = strings.TrimSuffix(strings.TrimSuffix(base, ".yaml"), ".yml")
	name := strcase.ToKebab(base)
	request := domain.Request{
		Name:   name,
		Method: strings.ToUpper(method),
		URL:    "/",
	}
	return s.requestRepository.Save(ctx, root, requestPath, request)
}

// ensureCollections makes sure every folder along collectionPath has a
// collection.yml, creating any that are missing. It is idempotent — existing
// collections are left untouched.
func (s *requestService) ensureCollections(ctx context.Context, root, collectionPath string) error {
	segments := strings.Split(strings.Trim(collectionPath, "/"), "/")
	for i := range segments {
		sub := strings.Join(segments[:i+1], "/")
		exists, err := s.collectionRepository.Exists(ctx, root, sub)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		name := strcase.ToKebab(path.Base(sub))
		if err := s.collectionRepository.Save(ctx, root, sub, domain.Collection{Name: name}); err != nil {
			return err
		}
	}
	return nil
}
