package game

import (
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-client/world"
)

type State struct {
	FocusedObjectID uint32
	FocusedImage    ui.ElementI
	world           world.World
}

func (s *State) Hook(func(s *State)) {

}
