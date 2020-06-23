package world

import (
	"errors"
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/data"
	"github.com/chimera-rpg/go-common/network"
)

// World is a collection of all the current known client representations of the game world.
type World struct {
	client     *client.Client
	maps       map[data.StringID]*DynamicMap
	currentMap data.StringID
	animations map[uint32]Animation
	objects    map[uint32]*Object
}

// Init initializes the given world object with the passed client.
func (w *World) Init(c *client.Client) {
	w.client = c
	w.maps = make(map[data.StringID]*DynamicMap)
	w.objects = make(map[uint32]*Object)
	w.animations = make(map[uint32]Animation)
	w.currentMap = 0
}

func (w *World) HandleMapCommand(cmd network.CommandMap) error {
	if _, ok := w.maps[cmd.MapID]; ok {
		// TODO: ?
	} else {
		w.client.Log.Printf("Made map %d(%s)\n", cmd.MapID, cmd.Name)
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
		// If animation id is not known, add the animation, then send an animation request.
		if _, animExists := w.animations[p.AnimationID]; !animExists {
			w.animations[p.AnimationID] = Animation{
				Faces: make(map[uint32][]AnimationFrame),
			}
			w.client.Log.Printf("Sending animation request for %d\n", p.AnimationID)
			w.client.Send(network.CommandAnimation{
				Type:        network.Get,
				AnimationID: p.AnimationID,
			})
		}
		// TODO: Check if AnimationID is known.
		w.CreateObjectFromPayload(cmd.ObjectID, p)
	case network.CommandObjectPayloadDelete:
		w.DeleteObject(cmd.ObjectID)
	default:
		w.client.Log.Printf("Unhandled CommandObject Payload: %+v\n", p)
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

func (w *World) HandleAnimationCommand(cmd network.CommandAnimation) error {
	if _, exists := w.animations[cmd.AnimationID]; !exists {
		w.animations[cmd.AnimationID] = Animation{
			AnimationID: cmd.AnimationID,
			Faces:       make(map[uint32][]AnimationFrame),
		}
	}
	for faceID, frames := range cmd.Faces {
		w.animations[cmd.AnimationID].Faces[faceID] = make([]AnimationFrame, len(frames))
		for frameIndex, frame := range frames {
			w.animations[cmd.AnimationID].Faces[faceID][frameIndex] = AnimationFrame{
				ImageID: frame.ImageID,
				Time:    frame.Time,
			}
		}
	}
	return nil
}
