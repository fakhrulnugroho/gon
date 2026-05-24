package driven

import (
	"gon/hexagonal/core/port/driving"
)

type RegistryService interface {
	Register(cmd driving.Command)
	List() []driving.Command
	Get(name string) (driving.Command, bool)
}
