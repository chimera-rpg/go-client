package world

import (
	"errors"

	"github.com/chimera-rpg/go-client/data"
	"github.com/chimera-rpg/go-common/network"
	"github.com/sirupsen/logrus"
)

// World is a collection of all the current known client representations of the game world.
type World struct {
	dataManager *data.Manager
	maps        map[data.StringID]*DynamicMap
	currentMap  data.StringID
	objects     map[uint32]*Object
	Log         *logrus.Logger
}

// Init initializes the given world object with the passed client.
func (w *World) Init(manager *data.Manager, l *logrus.Logger) {
	w.dataManager = manager
	w.Log = l

	w.maps = make(map[data.StringID]*DynamicMap)
	w.objects = make(map[uint32]*Object)
	w.currentMap = 0
}

// HandleMapCommand handles a map command, creating a new DynamicMap if it does not exist.
func (w *World) HandleMapCommand(cmd network.CommandMap) error {
	if _, ok := w.maps[cmd.MapID]; ok {
		// TODO: ?
	} else {
		w.Log.WithFields(logrus.Fields{
			"ID":   cmd.MapID,
			"Name": cmd.Name,
		}).Info("[World] Created map")
		w.maps[cmd.MapID] = &DynamicMap{
			height: uint32(cmd.Height),
			width:  uint32(cmd.Width),
			depth:  uint32(cmd.Depth),
		}
		w.maps[cmd.MapID].Init()
	}
	w.currentMap = cmd.MapID
	return nil
}

// HandleTileCommand handles a CommandTile, creating missing objects, updating object positions, and invalidates objects that go missing.
func (w *World) HandleTileCommand(cmd network.CommandTile) error {
	if _, ok := w.maps[w.currentMap]; !ok {
		return errors.New("cannot set tile, as no map exists")
	}
	// Create object if it does not exist and update its properties to match the tile coordinates.
	for oI, oID := range cmd.ObjectIDs {
		if _, ok := w.objects[oID]; !ok {
			w.objects[oID] = &Object{}
		} else {
			if w.objects[oID].Y != cmd.Y || w.objects[oID].X != cmd.X || w.objects[oID].Z != cmd.Z || w.objects[oID].Index != oI {
				w.objects[oID].Changed = true
			}
		}
		w.objects[oID].Y = cmd.Y
		w.objects[oID].X = cmd.X
		w.objects[oID].Z = cmd.Z
		w.objects[oID].Index = oI
		w.objects[oID].Missing = false
	}
	// See if we need to invalidate any objects that no longer are contained in the given tile.
	for _, oID := range w.maps[w.currentMap].GetTile(cmd.Y, cmd.X, cmd.Z).objectIDs {
		if _, ok := w.objects[oID]; !ok {
			continue
		}
		stillExists := false
		for _, newID := range cmd.ObjectIDs {
			if newID == oID {
				stillExists = true
				break
			}
		}
		// If the tile does not exist here _AND_ the object is still marked as being here, then flag the object as missing.
		if !stillExists {
			if w.objects[oID].Y == cmd.Y && w.objects[oID].X == cmd.X && w.objects[oID].Z == cmd.Z {
				w.objects[oID].Missing = true
			}
		}
	}
	// Set the map tile.
	w.maps[w.currentMap].SetTile(cmd.Y, cmd.X, cmd.Z, cmd.ObjectIDs)
	return nil
}

// HandleObjectCommand handles an ObjectCommand, creating or deleting depending on the payload.
func (w *World) HandleObjectCommand(cmd network.CommandObject) error {
	switch p := cmd.Payload.(type) {
	case network.CommandObjectPayloadCreate:
		w.dataManager.EnsureAnimation(p.AnimationID)
		w.CreateObjectFromPayload(cmd.ObjectID, p)
	case network.CommandObjectPayloadDelete:
		w.DeleteObject(cmd.ObjectID)
	default:
		w.Log.WithFields(logrus.Fields{
			"payload": p,
		}).Info("[World] Unhandled CommandObject Payload")
	}
	return nil
}

// CreateObjectFromPayload creates or updates an Object associated with an object ID from a creation payload.
func (w *World) CreateObjectFromPayload(oID uint32, p network.CommandObjectPayloadCreate) error {
	if _, ok := w.objects[oID]; ok {
		// Update existing object.
		w.objects[oID].Type = p.TypeID
		w.objects[oID].AnimationID = p.AnimationID
		w.objects[oID].FaceID = p.FaceID
	} else {
		// Create a new object.
		w.objects[oID] = &Object{
			ID:          oID,
			Type:        p.TypeID,
			AnimationID: p.AnimationID,
			FaceID:      p.FaceID,
			Missing:     true,
			H:           p.Height,
			W:           p.Width,
			D:           p.Depth,
		}
	}
	return nil
}

// DeleteObject deletes the given object ID from the world's objects field.
func (w *World) DeleteObject(oID uint32) error {
	delete(w.objects, oID)
	return nil
}

// GetObjects returns an array of all objects the client knows about.
func (w *World) GetObjects() []*Object {
	objects := make([]*Object, len(w.objects))
	oI := 0
	for _, o := range w.objects {
		objects[oI] = o
		oI++
	}
	return objects
}

// GetObject returns a pointer to an object based upon its ID.
func (w *World) GetObject(oID uint32) *Object {
	return w.objects[oID]
}

// GetCurrentMap returns a pointer to the current map.
func (w *World) GetCurrentMap() *DynamicMap {
	return w.maps[w.currentMap]
}
