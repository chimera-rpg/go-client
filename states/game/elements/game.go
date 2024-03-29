package elements

import (
	"github.com/chimera-rpg/go-client/config"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-client/world"
	"github.com/chimera-rpg/go-server/network"
)

// game is a common interface for elements to access the Game's properties.
type game interface {
	World() *world.World
	FocusedImage() ui.ElementI
	FocusedObjectID() uint32
	FocusedObject() *world.Object
	FocusObject(uint32)
	HookEvent(interface{}, func(e interface{}))
	InputChan() chan interface{}
	SendNetMessage(network.Command)
	Config() *config.Config
	Styles() map[string]map[string]string
	Slot(uint32) string
	TypeHint(uint32) string
}
