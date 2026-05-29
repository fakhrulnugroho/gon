package domain

type Config struct {
	Path    string
	Query   map[string]string
	Headers map[string]string
	Body    string
}
