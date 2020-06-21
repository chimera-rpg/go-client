package world

import (
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/data"
	"github.com/chimera-rpg/go-common/network"

	"fmt"
)

// World is a collection of all the current known client representations of the game world.
type World struct {
	client     *client.Client
	maps       map[data.StringID]DynamicMap
	currentMap data.StringID
}

// Init initializes the given world object with the passed client.
func (w *World) Init(c *client.Client) {
	w.client = c
	w.maps = make(map[data.StringID]DynamicMap)
	w.currentMap = 0
}

// HandleNet is the handler for all network updates.
func (w *World) HandleNet(cmd network.Command) {
	// TODO: process commands
	switch cmd.(type) {
	case network.CommandMap:
		fmt.Printf("Got map command: %+v\n", cmd)
	case network.CommandObject:
		fmt.Printf("Got object command: %+v\n", cmd)
	case network.CommandTile:
		fmt.Printf("Got tile command: %+v\n", cmd)
	}
}
