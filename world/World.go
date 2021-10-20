package world

import (
	"errors"
	"math"

	"github.com/chimera-rpg/go-client/data"
	"github.com/chimera-rpg/go-common/network"
	"github.com/sirupsen/logrus"
)

// World is a collection of all the current known client representations of the game world.
type World struct {
	dataManager  *data.Manager
	maps         map[data.StringID]*DynamicMap
	currentMap   data.StringID
	objects      map[uint32]*Object
	viewObjectID uint32
	visibleTiles map[TileKey]struct{}
	Log          *logrus.Logger
}

// Init initializes the given world object with the passed client.
func (w *World) Init(manager *data.Manager, l *logrus.Logger) {
	w.dataManager = manager
	w.Log = l

	w.maps = make(map[data.StringID]*DynamicMap)
	w.objects = make(map[uint32]*Object)
	w.visibleTiles = make(map[TileKey]struct{})
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
	viewChanged := false
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
		if oID == w.viewObjectID {
			viewChanged = true
		}
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

	// Update our visible tiles if the view object moved.
	if viewChanged {
		w.updateVisibleTiles()
	}
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
	case network.CommandObjectPayloadViewTarget:
		w.viewObjectID = cmd.ObjectID
		w.updateVisibleTiles()
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
			Opaque:      p.Opaque,
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

// GetViewObject returns a pointer to the object which the view should be centered on.
func (w *World) GetViewObject() *Object {
	return w.objects[w.viewObjectID]
}

// GetCurrentMap returns a pointer to the current map.
func (w *World) GetCurrentMap() *DynamicMap {
	return w.maps[w.currentMap]
}

// HandleNoiseCommand handles CommandNoise.
func (w *World) HandleNoiseCommand(cmd network.CommandNoise) error {
	w.dataManager.EnsureAudio(cmd.AudioID)
	return nil
}

func (w *World) getVisibilitySphere(radius float64) (targets [][3]float64) {
	stackCount := 20
	sliceCount := 20
	for stack := 0; stack < stackCount-1; stack++ {
		phi := math.Pi * float64(stack+1) / float64(stackCount)
		for slice := 0; slice < sliceCount; slice++ {
			theta := 2.0 * math.Pi * float64(slice) / float64(sliceCount)
			y := math.Cos(phi)
			x := math.Sin(phi) * math.Cos(theta)
			z := math.Sin(phi) * math.Sin(theta)
			targets = append(targets, [3]float64{y * radius, x * radius, z * radius})
		}
	}
	return targets
}

func (w *World) getVisibilityCube(yi, xi, zi int, height, width, depth int) (c [][3]float64) {
	vhh := height / 2
	vwh := width / 2
	vdh := depth / 2

	m := w.GetCurrentMap()
	// TODO: Use target object's statistics for vision range.

	ymin := yi - vhh
	if ymin < 0 {
		ymin = 0
	}
	ymax := yi + vhh
	if ymax > int(m.GetHeight()) {
		ymax = int(m.GetHeight()) - 1
	}

	xmin := xi - vwh
	if xmin < 0 {
		xmin = 0
	}
	xmax := xi + vwh
	if xmax > int(m.GetWidth()) {
		xmax = int(m.GetWidth()) - 1
	}

	zmin := zi - vdh
	if zmin < 0 {
		zmin = 0
	}
	zmax := zi + vdh
	if zmax > int(m.GetDepth()) {
		zmax = int(m.GetDepth()) - 1
	}

	// massive cube
	for y := ymin; y < ymax; y++ {
		for x := xmin; x < xmax; x++ {
			for z := zmin; z < zmax; z++ {
				c = append(c, [3]float64{float64(y), float64(x), float64(z)})
			}
		}
	}

	return c
}

// Line of sight stuff
func (w *World) updateVisibleTiles() {
	visibleTiles := make(map[TileKey]struct{})
	// Collect our box for rays
	o := w.GetViewObject()
	if o == nil {
		return
	}

	rayEnds := w.getVisibilityCube(int(o.Y)+int(o.H), int(o.X), int(o.Z), 16, 32, 32)

	//rayEnds := w.getVisibilitySphere(20)

	// Now let's shoot some rays via Amanatides & Woo.
	m := w.GetCurrentMap()
	y1 := float64(int(o.Y) + int(o.H))
	if y1 >= float64(m.GetHeight()) {
		y1 = float64(m.GetHeight() - 1)
	}
	x1 := float64(o.X)
	z1 := float64(o.Z + 1)

	for _, v := range rayEnds {
		var tMaxX, tMaxY, tMaxZ, tDeltaX, tDeltaY, tDeltaZ float64
		y2 := v[0]
		x2 := v[1]
		z2 := v[2]
		var dy, dx, dz int
		var y, x, z float64

		sign := func(x float64) int {
			if x > 0 {
				return 1
			} else if x < 0 {
				return -1
			}
			return 0
		}
		frac0 := func(x float64) float64 {
			return x - math.Floor(x)
		}
		frac1 := func(x float64) float64 {
			return 1 - x + math.Floor(x)
		}

		dy = sign(y2 - y1)
		if dy != 0 {
			tDeltaY = math.Min(float64(dy)/(y2-y1), 1000000)
		} else {
			tDeltaY = 1000000
		}
		if dy > 0 {
			tMaxY = tDeltaY * frac1(y1)
		} else {
			tMaxY = tDeltaY * frac0(y1)
		}
		y = y1

		dx = sign(x2 - x1)
		if dx != 0 {
			tDeltaX = math.Min(float64(dx)/(x2-x1), 1000000)
		} else {
			tDeltaX = 1000000
		}
		if dx > 0 {
			tMaxX = tDeltaX * frac1(x1)
		} else {
			tMaxX = tDeltaX * frac0(x1)
		}
		x = x1

		dz = sign(z2 - z1)
		if dz != 0 {
			tDeltaZ = math.Min(float64(dz)/(z2-z1), 1000000)
		} else {
			tDeltaZ = 1000000
		}
		if dz > 0 {
			tMaxZ = tDeltaZ * frac1(z1)
		} else {
			tMaxZ = tDeltaZ * frac0(z1)
		}
		z = z1

		for {
			if tMaxX < tMaxY {
				if tMaxX < tMaxZ {
					x += float64(dx)
					tMaxX += tDeltaX
				} else {
					z += float64(dz)
					tMaxZ += tDeltaZ
				}
			} else {
				if tMaxY < tMaxZ {
					y += float64(dy)
					tMaxY += tDeltaY
				} else {
					z += float64(dz)
					tMaxZ += tDeltaZ
				}
			}
			if tMaxY > 1 && tMaxX > 1 && tMaxZ > 1 {
				break
			}
			if y < 0 || x < 0 || z < 0 || y >= float64(m.GetHeight()) || x >= float64(m.GetWidth()) || z >= float64(m.GetDepth()) {
				continue
			}
			tile := m.GetTile(uint32(y), uint32(x), uint32(z))
			opaque := false
			for _, oID := range tile.GetObjects() {
				o := w.GetObject(oID)
				if o == nil {
					continue
				}
				if o.Opaque {
					opaque = true
					break
				}
			}
			visibleTiles[TileKey{Y: uint32(y), X: uint32(x), Z: uint32(z)}] = struct{}{}

			if opaque {
				break
			}
		}
	}

	// Set objects no longer visible
	for tk := range w.visibleTiles {
		_, isVisible := visibleTiles[tk]
		if tiles, ok := m.tiles[tk]; ok {
			for _, oID := range tiles.objectIDs {
				if o, ok := w.objects[oID]; ok {
					if !isVisible && o.Visible {
						o.Visible = false
						o.VisibilityChange = true
					}
				}
			}
		}
	}

	for tk := range visibleTiles {
		if tiles, ok := m.tiles[tk]; ok {
			for _, oID := range tiles.objectIDs {
				if o, ok := w.objects[oID]; ok {
					if !o.Visible {
						o.Visible = true
						o.VisibilityChange = true
					}
				}
			}
		}
	}

	w.visibleTiles = visibleTiles
}
