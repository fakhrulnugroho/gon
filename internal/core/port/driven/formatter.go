package driven

type Formatter[T any] interface {
	Format(t T) string
}
