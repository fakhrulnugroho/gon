package driving

import "context"

type WorkspaceService interface {
	Create(ctx context.Context, directory string) error
}
