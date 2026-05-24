package driving

import "context"

type Command interface {
	Name() string
	Description() string
	Execute(ctx context.Context, args []string) error
}
