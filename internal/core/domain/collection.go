package domain

// Collection is a folder of requests sharing a Config. The Config defaults are
// inherited by every request in the folder (and, for nested folders, by child
// collections too).
type Collection struct {
	Name        string
	Description string
	Config      Config
}
