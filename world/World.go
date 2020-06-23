package world

import (
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/data"
	"github.com/chimera-rpg/go-common/network"
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

func (w *World) HandleMapCommand(cmd network.CommandMap) error {
	if _, ok := w.maps[cmd.MapID]; ok {
		// TODO: ?
	} else {
		w.maps[cmd.MapID] = DynamicMap{}
	}
	return nil
}

func (w *World) HandleObjectCommand(cmd network.CommandObject) error {
	switch p := cmd.Payload.(type) {
	case network.CommandObjectPayloadCreate:
		w.client.Log.Printf("Got CommandObjectPayloadCreate: %+v\n", p)
		// TODO: Check if AnimationID is known.
		// TODO: Add object representation to map.
	default:
		w.client.Log.Printf("Unhandled CommandObject Payload: %+v\n", p)
	}
	return nil
}

func (w *World) HandleTileCommand(cmd network.CommandTile) error {
	return nil
}
