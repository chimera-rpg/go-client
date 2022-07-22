package world

import (
	"errors"
	"math"
	"math/rand"

	"github.com/chimera-rpg/go-client/data"
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-common/network"
	"github.com/sirupsen/logrus"
)

// World is a collection of all the current known client representations of the game world.
type World struct {
	dataManager             *data.Manager
	maps                    map[data.StringID]*DynamicMap
	currentMap              data.StringID
	objects                 []*Object
	PendingObjectAnimations map[data.StringID][]uint32 // Map of animations to objects waiting for their animation exist.
	viewObjectID            uint32
	viewObject              *Object
	deletedObjects          []uint32 // A list of deleted object IDs. Used and cleared during the render call.
	visibleTiles            [][][]bool
	unblockedTiles          [][][]bool
	Log                     *logrus.Logger
}

// Init initializes the given world object with the passed client.
func (w *World) Init(manager *data.Manager, l *logrus.Logger) {
	w.dataManager = manager
	w.Log = l

	w.maps = make(map[data.StringID]*DynamicMap)
	w.objects = make([]*Object, 0)
	w.visibleTiles = make([][][]bool, 0)
	w.unblockedTiles = make([][][]bool, 0)
	w.PendingObjectAnimations = make(map[uint32][]uint32)
	w.currentMap = 0
}

// HandleMapCommand handles a map command, creating a new DynamicMap if it does not exist.
func (w *World) HandleMapCommand(cmd network.CommandMap) error {
	// TODO: We have this multiple map code because in the future I wanted to have maps able to be tiled together.
	/*	if _, ok := w.maps[cmd.MapID]; ok {
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
		}*/
	w.maps[cmd.MapID] = &DynamicMap{
		height: uint32(cmd.Height),
		width:  uint32(cmd.Width),
		depth:  uint32(cmd.Depth),
	}
	w.maps[cmd.MapID].Init()

	w.currentMap = cmd.MapID

	// Clear out our known objects. This should really be managed differently, somehow.
	p := w.GetViewObject()
	for _, o := range w.objects {
		if t := w.GetCurrentMap().GetTile(int(o.Y), int(o.X), int(o.Z)); t != nil {
			t.RemoveObject(o)
		}
	}
	w.objects = make([]*Object, 0)

	// Restore our known visible object if we have one.
	if p != nil {
		w.AddObject(p)
	}

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
		o := w.GetObject(oID)
		if o == nil {
			w.AddObject(&Object{
				ID: oID,
			})
		} else {
			if o.Y != cmd.Y || o.X != cmd.X || o.Z != cmd.Z || o.Index != oI {
				o.Changed = true
			}
		}
		o.Y = cmd.Y
		o.X = cmd.X
		o.Z = cmd.Z
		o.Index = oI
		o.Missing = false
		if oID == w.viewObjectID {
			w.viewObject = o
			viewChanged = true
		}
	}
	// See if we need to invalidate any objects that no longer are contained in the given tile.
	for _, tileObject := range w.maps[w.currentMap].GetTile(int(cmd.Y), int(cmd.X), int(cmd.Z)).objects {
		o := w.GetObject(uint32(tileObject.ID))
		if o == nil {
			continue
		}
		stillExists := false
		for _, newID := range cmd.ObjectIDs {
			if newID == tileObject.ID {
				stillExists = true
				break
			}
		}
		// If the tile does not exist here _AND_ the object is still marked as being here, then flag the object as missing.
		if !stillExists {
			if o.Y == cmd.Y && o.X == cmd.X && o.Z == cmd.Z {
				o.Missing = true
			}
		}
	}
	// Set the map tile.
	var objects []*Object
	for _, oID := range cmd.ObjectIDs {
		if o := w.GetObject(oID); o != nil {
			objects = append(objects, o)
		}
	}
	w.maps[w.currentMap].SetTile(cmd.Y, cmd.X, cmd.Z, objects)

	// Update our visible tiles if the view object moved.
	if viewChanged {
		w.updateVisibleTiles()
		w.updateVisionUnblocking()
	}
	return nil
}

func (w *World) HandleTileLightCommand(cmd network.CommandTileLight) error {
	if _, ok := w.maps[w.currentMap]; !ok {
		return errors.New("cannot set tile light, as no map exists")
	}
	w.maps[w.currentMap].SetTileLight(cmd.Y, cmd.X, cmd.Z, cmd.Brightness)
	t := w.maps[w.currentMap].GetTile(int(cmd.Y), int(cmd.X), int(cmd.Z))
	if t != nil {
		for _, o := range t.objects {
			o.LightingChange = true
			o.Brightness = t.brightness
		}
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
		w.viewObject = w.GetObject(cmd.ObjectID)
		w.updateVisibleTiles()
		w.updateVisionUnblocking()
	default:
		w.Log.WithFields(logrus.Fields{
			"payload": p,
		}).Info("[World] Unhandled CommandObject Payload")
	}
	return nil
}

// CreateObjectFromPayload creates or updates an Object associated with an object ID from a creation payload.
func (w *World) CreateObjectFromPayload(oID uint32, p network.CommandObjectPayloadCreate) error {
	o := w.GetObject(oID)
	if o != nil {
		// Update existing object.
		o.Type = p.TypeID

		if o.AnimationID != p.AnimationID {
			// Get randomized frame start if we have the associated animation.
			if anim := w.dataManager.GetAnimation(p.AnimationID); anim.Ready {
				face := anim.GetFace(p.FaceID)
				o.Animation = anim
				o.Face = face
				if anim.RandomFrame {
					o.FrameIndex = rand.Intn(len(face.Frames))
				}
				o.ImageChanged = true
			} else {
				// Animation does not yet exist, add it to the pending.
				w.PendingObjectAnimations[p.AnimationID] = append(w.PendingObjectAnimations[p.AnimationID], oID)
			}
		}
		o.AnimationID = p.AnimationID
		o.FaceID = p.FaceID
	} else {
		// Create a new object.
		o = &Object{
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
		// Get randomized frame start if we have the associated animation.
		if anim := w.dataManager.GetAnimation(p.AnimationID); anim.Ready {
			face := anim.GetFace(p.FaceID)
			o.Animation = anim
			o.Face = face
			if anim.RandomFrame {
				o.FrameIndex = rand.Intn(len(face.Frames))
			}
			o.ImageChanged = true
		} else {
			// Animation does not yet exist, add it to the pending.
			w.PendingObjectAnimations[p.AnimationID] = append(w.PendingObjectAnimations[p.AnimationID], oID)
		}
		w.AddObject(o)
	}

	// Ensure our shadow gets created.
	if o.Type == cdata.ArchetypeNPC.AsUint8() || o.Type == cdata.ArchetypePC.AsUint8() || o.Type == cdata.ArchetypeItem.AsUint8() {
		o.HasShadow = true
	}

	return nil
}

// AddObject adds the given object to the objects slice.
func (w *World) AddObject(o *Object) {
	w.objects = append(w.objects, o)
}

// DeleteObject deletes the given object ID from the world's objects field.
func (w *World) DeleteObject(oID uint32) error {
	o := w.GetObject(oID)
	if o != nil {
		if t := w.GetCurrentMap().GetTile(int(o.Y), int(o.X), int(o.Z)); t != nil {
			t.RemoveObject(o)
		}
		for i, o2 := range w.objects {
			if o2.ID == oID {
				w.objects = append(w.objects[:i], w.objects[i+1:]...)
				break
			}
		}
		// Also remove the element since we moved elements to be part of objects directly.
		if o.Element != nil {
			o.Element.GetDestroyChannel() <- true
			o.Element = nil
		}
		// Remove shadow element.
		if o.ShadowElement != nil {
			o.ShadowElement.GetDestroyChannel() <- true
			o.ShadowElement = nil
		}
	}
	//w.deletedObjects = append(w.deletedObjects, oID)
	return nil
}

// GetDeletedObjects returns the deleted objects list.
func (w *World) GetDeletedObjects() []uint32 {
	return w.deletedObjects
}

// ClearDeleteObjects clears the deleted objects list.
func (w *World) ClearDeletedObjects() {
	for _, oID := range w.deletedObjects {
		// Remove from owning tile.
		if o := w.GetObject(oID); o != nil {
			t := w.GetCurrentMap().GetTile(int(o.Y), int(o.X), int(o.Z))
			t.RemoveObject(o)
		}

		for i, o2 := range w.objects {
			if o2.ID == oID {
				w.objects = append(w.objects[:i], w.objects[i+1:]...)
				break
			}
		}
	}
	w.deletedObjects = make([]uint32, 0)
}

// GetObjects returns an array of all objects the client knows about.
func (w *World) GetObjects() []*Object {
	return w.objects
}

// GetObject returns a pointer to an object based upon its ID.
func (w *World) GetObject(oID uint32) *Object {
	for _, o := range w.objects {
		if o.ID == oID {
			return o
		}
	}
	return nil
}

// GetViewObject returns a pointer to the object which the view should be centered on.
func (w *World) GetViewObject() *Object {
	return w.viewObject
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

func (w *World) getSphereRays(yi, xi, zi int, radius float64) (targets [][2][3]float64) {
	stackCount := 20
	sliceCount := 20
	for stack := 0; stack < stackCount-1; stack++ {
		phi := math.Pi * float64(stack+1) / float64(stackCount)
		for slice := 0; slice < sliceCount; slice++ {
			theta := 2.0 * math.Pi * float64(slice) / float64(sliceCount)
			y := math.Cos(phi)
			x := math.Sin(phi) * math.Cos(theta)
			z := math.Sin(phi) * math.Sin(theta)
			targets = append(targets, [2][3]float64{{float64(yi), float64(xi), float64(zi)}, {float64(yi) + y*radius, float64(xi) + x*radius, float64(zi) + z*radius}})
		}
	}
	return targets
}

func (w *World) getCubeRays(originY, originX, originZ float64, minY, minX, minZ, maxY, maxX, maxZ int) (c [][2][3]float64) {
	for y := minY; y < maxX; y++ {
		for x := minX; x < maxX; x++ {
			for z := minZ; z < maxZ; z++ {
				c = append(c, [2][3]float64{{originY, originX, originZ}, {float64(y), float64(x), float64(z)}})
			}
		}
	}
	return c
}

func (w *World) rayCasts(rays [][2][3]float64, maxY, maxX, maxZ float64, hit func(y, x, z int) bool) {
	for _, v := range rays {
		var tMaxX, tMaxY, tMaxZ, tDeltaX, tDeltaY, tDeltaZ float64
		y1 := v[0][0]
		x1 := v[0][1]
		z1 := v[0][2]
		y2 := v[1][0]
		x2 := v[1][1]
		z2 := v[1][2]
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
			if y < 0 || x < 0 || z < 0 || y >= maxY || x >= maxX || z >= maxZ {
				continue
			}

			if hit(int(y), int(x), int(z)) {
				break
			}
		}
	}
}

// Line of sight stuff
func (w *World) updateVisibleTiles() {
	o := w.GetViewObject()
	if o == nil {
		return
	}

	// 1. Collect our rays
	m := w.GetCurrentMap()
	y1 := float64(int(o.Y) + int(o.H))
	if y1 >= float64(m.GetHeight()) {
		y1 = float64(m.GetHeight() - 1)
	}
	x1 := float64(o.X)
	z1 := float64(o.Z)

	// Acquire our box dimensions
	vhh := float64(32 / 2)
	vwh := float64(32 / 2)
	vdh := float64(32 / 2)

	ymin := y1 - vhh
	if ymin < 0 {
		ymin = 0
	}
	ymax := y1 + vhh
	if ymax > float64(m.GetHeight()) {
		ymax = float64(m.GetHeight()) - 1
	}

	xmin := x1 - vwh
	if xmin < 0 {
		xmin = 0
	}
	xmax := x1 + vwh
	if xmax > float64(m.GetWidth()) {
		xmax = float64(m.GetWidth()) - 1
	}

	zmin := z1 - vdh
	if zmin < 0 {
		zmin = 0
	}
	zmax := z1 + vdh
	if zmax > float64(m.GetDepth()) {
		zmax = float64(m.GetDepth()) - 1
	}

	rays := w.getCubeRays(y1, x1, z1, int(ymin), int(xmin), int(zmin), int(ymax), int(xmax), int(zmax))

	visibleTiles := make([][][]bool, m.GetHeight())
	for i := range visibleTiles {
		visibleTiles[i] = make([][]bool, m.GetWidth())
		for j := range visibleTiles[i] {
			visibleTiles[i][j] = make([]bool, m.GetDepth())
		}
	}

	markTiles := func(y, x, z int) bool {
		visibleTiles[y][x][z] = true

		tile := m.GetTile(y, x, z)

		for _, o := range tile.objects {
			if o.Opaque {
				return true
			}
		}
		return false
	}

	// This feels wrong, but we duplicate the rays and offset the origin to ensure we can see over vertical edges on character's sides.
	for _, r := range rays {
		rays = append(rays, [2][3]float64{
			{r[0][0], r[0][1] + float64(o.W), r[0][2] + float64(o.D) + 1},
			{r[1][0], r[1][1], r[1][2]},
		})
	}

	// Now let's shoot some rays via Amanatides & Woo.
	w.rayCasts(rays, float64(m.GetHeight()), float64(m.GetWidth()), float64(m.GetDepth()), markTiles)
	// Set objects no longer visible
	for y := range visibleTiles {
		for x := range visibleTiles[y] {
			for z := range visibleTiles[y][x] {
				isVisible := visibleTiles[y][x][z]
				tiles := m.GetTile(y, x, z)
				for _, o := range tiles.objects {
					if !isVisible && o.Visible {
						o.Visible = false
						o.VisibilityChange = true
					} else if isVisible && !o.Visible {
						o.Visible = true
						o.VisibilityChange = true
					}
				}
			}
		}
	}

	w.visibleTiles = visibleTiles
}

func (w *World) updateVisionUnblocking() {
	o := w.GetViewObject()
	if o == nil {
		return
	}
	m := w.GetCurrentMap()

	// Collect our end-points for rays
	oY := float64(o.Y) + float64(o.H)/2
	oX := float64(o.X) + float64(o.W)/2
	oZ := float64(o.Z)

	minY := int(oY) + 6
	maxY := minY + int(o.H) + 8
	minX := int(oX) - 4
	maxX := minX + int(o.W) + 6
	minZ := int(oZ) + 3
	maxZ := minZ + int(o.D) + 8

	rays := w.getCubeRays(oY, oX, oZ, minY, minX, minZ, maxY, maxX, maxZ)
	// TODO: We actually need to use an angled cone, originating from the near view target origin to whatever area we deem as the "camera" area
	// TODO: Or, we could have 2 "cubes" -- basically 2 flat cubes that create a "right angle bracket"

	unblockedTiles := make([][][]bool, m.GetHeight())
	for i := range unblockedTiles {
		unblockedTiles[i] = make([][]bool, m.GetWidth())
		for j := range unblockedTiles[i] {
			unblockedTiles[i][j] = make([]bool, m.GetDepth())
		}
	}

	// Now let's shoot some rays via Amanatides & Woo.
	w.rayCasts(rays, float64(m.GetHeight()), float64(m.GetWidth()), float64(m.GetDepth()), func(y, x, z int) bool {
		t := m.GetTile(y, x, z)
		opaque := false
		for _, o := range t.objects {
			if o.Opaque {
				opaque = true
			}
		}
		if opaque {
			unblockedTiles[y][x][z] = true
		}
		return false
	})

	// Set objects no longer Unblocked
	for y := range unblockedTiles {
		for x := range unblockedTiles[y] {
			for z := range unblockedTiles[y][x] {
				isUnblocked := unblockedTiles[y][x][z]
				tiles := m.GetTile(y, x, z)
				for _, o := range tiles.objects {
					if !isUnblocked && o.Unblocked {
						o.Unblocked = false
						o.UnblockedChange = true
					} else if isUnblocked && !o.Unblocked {
						o.Unblocked = true
						o.UnblockedChange = true
					}
				}
			}
		}
	}

	w.unblockedTiles = unblockedTiles
}

// GetObjectShadowPosition returns the shadow position for the given object. This is calculated from the object's position downward (-Y) until an opaque block is eached.
func (w *World) GetObjectShadowPosition(o *Object) (y, x, z int) {
	return
	y = int(o.Y)
	x = int(o.X)
	z = int(o.Z)

	for i := y; i > 0; i-- {
		for _, o2 := range w.maps[w.currentMap].GetTile(i, x, z).objects {
			if o2.Opaque {
				y = i + 1
				// If it's an opaque tile, we treat its shadow position as one lower.
				if o2.Type == cdata.ArchetypeTile.AsUint8() {
					y--
				}
				return
			}
		}
	}

	return
}

func (w *World) CheckPendingObjectAnimations(animationID uint32) {
	if pending, ok := w.PendingObjectAnimations[animationID]; ok {
		anim := w.dataManager.GetAnimation(animationID)
		for _, objectID := range pending {
			if o := w.GetObject(objectID); o != nil {
				face := anim.GetFace(o.FaceID)
				if anim.RandomFrame {
					o.FrameIndex = rand.Intn(len(face.Frames))
				}
				o.Animation = anim
				o.Face = face
				o.ImageChanged = true
			}
		}
		delete(w.PendingObjectAnimations, animationID)
	}
}
