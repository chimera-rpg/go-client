package world

import (
	"errors"
	"github.com/chimera-rpg/go-client/data"
	"github.com/chimera-rpg/go-common/network"
	"log"
	"os"
)

// World is a collection of all the current known client representations of the game world.
type World struct {
	dataManager *data.Manager
	maps        map[data.StringID]*DynamicMap
	currentMap  data.StringID
	objects     map[uint32]*Object
	Log         *log.Logger
}

// Init initializes the given world object with the passed client.
func (w *World) Init(manager *data.Manager) {
	w.dataManager = manager
	w.Log = log.New(os.Stdout, "World: ", log.Ltime)
	w.maps = make(map[data.StringID]*DynamicMap)
	w.objects = make(map[uint32]*Object)
	w.currentMap = 0
}

func (w *World) HandleMapCommand(cmd network.CommandMap) error {
	if _, ok := w.maps[cmd.MapID]; ok {
		// TODO: ?
	} else {
		w.Log.Printf("Made map %d(%s)\n", cmd.MapID, cmd.Name)
		w.maps[cmd.MapID] = &DynamicMap{}
		w.maps[cmd.MapID].Init()
	}
	w.currentMap = cmd.MapID
	return nil
}

func (w *World) HandleTileCommand(cmd network.CommandTile) error {
	if _, ok := w.maps[w.currentMap]; !ok {
		return errors.New("cannot set tile, as no map exists")
	}
	w.maps[w.currentMap].SetTile(cmd.Y, cmd.X, cmd.Z, cmd.ObjectIDs)
	return nil
}

func (w *World) HandleObjectCommand(cmd network.CommandObject) error {
	switch p := cmd.Payload.(type) {
	case network.CommandObjectPayloadCreate:
		w.dataManager.EnsureAnimation(p.AnimationID)
		w.CreateObjectFromPayload(cmd.ObjectID, p)
	case network.CommandObjectPayloadDelete:
		w.DeleteObject(cmd.ObjectID)
	default:
		w.Log.Printf("Unhandled CommandObject Payload: %+v\n", p)
	}
	return nil
}

func (w *World) CreateObjectFromPayload(oID uint32, p network.CommandObjectPayloadCreate) error {
	if _, ok := w.objects[oID]; ok {
		return errors.New("Object already exists...")
	}
	w.objects[oID] = &Object{
		ID:          oID,
		Type:        p.TypeID,
		AnimationID: p.AnimationID,
		FaceID:      p.FaceID,
	}
	return nil
}

func (w *World) DeleteObject(oID uint32) error {
	delete(w.objects, oID)
	return nil
}
