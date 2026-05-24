package service

import (
	"gon/hexagonal/core/port/driven"
	"gon/hexagonal/core/port/driving"
)

type registryService struct {
	commands map[string]driving.Command
}

func NewRegistryService() driven.RegistryService {
	return &registryService{
		commands: make(map[string]driving.Command),
	}
}

func (r *registryService) Register(cmd driving.Command) {
	r.commands[cmd.Name()] = cmd
}

func (r *registryService) Get(name string) (driving.Command, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

func (r *registryService) List() []driving.Command {
	cmds := make([]driving.Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		cmds = append(cmds, cmd)
	}
	return cmds
}
