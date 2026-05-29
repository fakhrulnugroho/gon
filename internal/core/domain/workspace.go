package domain

type Workspace struct {
	Name             string
	BaseURL          string
	WorkingDirectory string
	Config           Config
}
