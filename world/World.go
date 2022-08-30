package world

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"

	"github.com/chimera-rpg/go-client/data"
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-common/network"
	"github.com/sirupsen/logrus"
)

// World is a collection of all the current known client representations of the game world.
type World struct {
	dataManager                      *data.Manager
	maps                             map[data.StringID]*DynamicMap
	currentMap                       data.StringID
	objects                          []*Object
	changedObjects                   []*Object
	ReachCube                        [][][]struct{}
	IntersectCube                    [][][]struct{}
	PendingObjectAnimations          map[data.StringID][]uint32 // Map of animations to objects waiting for their animation exist.
	PendingObjectImages              map[uint32][]uint32        // Map of images to objects waiting.
	viewObjectID                     uint32
	viewObject                       *Object
	viewHeight, viewWidth, viewDepth int
	deletedObjects                   []uint32 // A list of deleted object IDs. Used and cleared during the render call.
	visibleTiles                     []bool
	unblockedTiles                   [][][]bool
	Log                              *logrus.Logger
	//
	LeftBlocked  bool
	FrontBlocked bool
	AboveBlocked bool
	InRoom       bool
}

// Init initializes the given world object with the passed client.
func (w *World) Init(manager *data.Manager, l *logrus.Logger) {
	w.dataManager = manager
	w.Log = l

	w.maps = make(map[data.StringID]*DynamicMap)
	w.objects = make([]*Object, 0)
	w.visibleTiles = make([]bool, 0)
	w.unblockedTiles = make([][][]bool, 0)
	w.PendingObjectAnimations = make(map[uint32][]uint32)
	w.PendingObjectImages = make(map[uint32][]uint32)
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
		height:            cmd.Height,
		width:             cmd.Width,
		depth:             cmd.Depth,
		outdoor:           cmd.Outdoor,
		outdoorBrightness: cmd.OutdoorBrightness,
		ambientHue:        cmd.AmbientHue,
		ambientBrightness: cmd.AmbientBrightness,
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

func (w *World) HandleTilesCommand(cmd network.CommandTiles) error {
	for _, t := range cmd.TileUpdates {
		w.HandleTileCommand(t)
	}
	for _, s := range cmd.SkyUpdates {
		w.HandleTileSkyCommand(s)
	}
	for _, l := range cmd.LightUpdates {
		w.HandleTileLightCommand(l)
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
			o = w.GetObject(oID)
		} else {
			if o.Y != int(cmd.Y) || o.X != int(cmd.X) || o.Z != int(cmd.Z) || o.Index != oI {
				o.Changed = true
			}
		}
		o.Y = int(cmd.Y)
		o.X = int(cmd.X)
		o.Z = int(cmd.Z)
		o.Index = oI
		o.Missing = false
		if oID == w.viewObjectID {
			w.viewObject = o
			viewChanged = true
		}
		w.changedObjects = append(w.changedObjects, o)
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
			if o.Y == int(cmd.Y) && o.X == int(cmd.X) && o.Z == int(cmd.Z) {
				o.Missing = true
				w.changedObjects = append(w.changedObjects, o)
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
	w.maps[w.currentMap].SetTile(int(cmd.Y), int(cmd.X), int(cmd.Z), objects)

	// Update our visible tiles if the view object moved.
	if viewChanged {
		w.updateVisibleTiles()
		w.updateVisionUnblocking()
	}
	w.updateTileLighting(int(cmd.Y), int(cmd.X), int(cmd.Z))
	return nil
}

func (w *World) updateTileLighting(y, x, z int) {
	t := w.maps[w.currentMap].GetTile(y, x, z)
	if t != nil {
		brightness := t.finalBrightness
		hue := t.finalHue
		/*if t.opaque {
			if t2 := w.maps[w.currentMap].GetTile(y+1, x, z); t2 != nil {
				if !t2.opaque {
					brightness = (brightness + t2.finalBrightness) / 2
					hue = (hue + t2.finalHue) / 2
				}
			}
		}*/

		for _, o := range t.objects {
			o.LightingChange = true
			o.Brightness = brightness
			o.Hue = hue
			o.Changed = true
			w.changedObjects = append(w.changedObjects, o)
		}
	}
}

func (w *World) HandleTileLightCommand(cmd network.CommandTileLight) error {
	if _, ok := w.maps[w.currentMap]; !ok {
		return errors.New("cannot set tile light, as no map exists")
	}
	w.maps[w.currentMap].SetTileLight(int(cmd.Y), int(cmd.X), int(cmd.Z), cmd.Brightness)
	w.maps[w.currentMap].SetTileHue(int(cmd.Y), int(cmd.X), int(cmd.Z), float32(cmd.Hue))
	w.maps[w.currentMap].RecalculateLightingAt(int(cmd.Y), int(cmd.X), int(cmd.Z))
	w.updateTileLighting(int(cmd.Y), int(cmd.X), int(cmd.Z))
	return nil
}

func (w *World) HandleTileSkyCommand(cmd network.CommandTileSky) error {
	if _, ok := w.maps[w.currentMap]; !ok {
		return errors.New("cannot set tile light, as no map exists")
	}
	w.maps[w.currentMap].SetTileSky(int(cmd.Y), int(cmd.X), int(cmd.Z), float32(cmd.Sky))
	w.maps[w.currentMap].RecalculateLightingAt(int(cmd.Y), int(cmd.X), int(cmd.Z))
	w.updateTileLighting(int(cmd.Y), int(cmd.X), int(cmd.Z))
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
		w.viewHeight = int(p.Height)
		w.viewWidth = int(p.Width)
		w.viewDepth = int(p.Depth)
		w.updateCubes()
		w.updateVisibleTiles()
		w.updateVisionUnblocking()
	case network.CommandObjectPayloadInfo:
		w.UpdateObjectInfo(cmd.ObjectID, p.Info)
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

		if o.AnimationID != p.AnimationID || o.FaceID != p.FaceID {
			// Get randomized frame start if we have the associated animation.
			if anim := w.dataManager.GetAnimation(p.AnimationID); anim.Ready {
				face := anim.GetFace(p.FaceID)
				o.Animation = anim
				o.Face = face
				if anim.RandomFrame {
					o.FrameIndex = rand.Intn(len(face.Frames))
				} else {
					o.FrameIndex = 0
				}
				o.Frame = &(face.Frames[o.FrameIndex])
			} else {
				// Animation does not yet exist, add it to the pending.
				w.PendingObjectAnimations[p.AnimationID] = append(w.PendingObjectAnimations[p.AnimationID], oID)
			}
			o.AnimationID = p.AnimationID
			o.FaceID = p.FaceID
			w.changedObjects = append(w.changedObjects, o)
		}
	} else {
		// Create a new object.
		o = &Object{
			ID:          oID,
			Type:        p.TypeID,
			AnimationID: p.AnimationID,
			FaceID:      p.FaceID,
			Missing:     true,
			H:           int8(p.Height),
			W:           int8(p.Width),
			D:           int8(p.Depth),
			Reach:       p.Reach,
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
			if o.FrameIndex < len(face.Frames) {
				o.Frame = &(face.Frames[o.FrameIndex])
			} else {
				fmt.Fprintf(os.Stderr, "missing frame %d\n", p.AnimationID)
			}
			// Check if we have an image to load.
			if o.Frame != nil {
				img, err := w.dataManager.GetCachedImage(o.Frame.ImageID)
				if err != nil {
					w.PendingObjectImages[o.Frame.ImageID] = append(w.PendingObjectImages[o.Frame.ImageID], oID)
				} else {
					o.Image = img
				}
			}
		} else {
			// Animation does not yet exist, add it to the pending.
			w.PendingObjectAnimations[p.AnimationID] = append(w.PendingObjectAnimations[p.AnimationID], oID)
		}
		w.AddObject(o)
		// FIXME: This is stupid to check this each time here. Make sure that when PayloadViewTarget is sent, the player's object is sent first.
		if w.viewObjectID == oID {
			w.viewObject = w.GetObject(oID)
			w.updateCubes()
			w.updateVisibleTiles()
			w.updateVisionUnblocking()
		}

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
	w.changedObjects = append(w.changedObjects, o)
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

func (w *World) UpdateObjectInfo(oID uint32, infos []cdata.ObjectInfo) {
	o := w.GetObject(oID)
	if o == nil {
		// Oops, somehow we got object info for something we don't know about.
		return
	}
	// TODO: Probably some sort of intelligent/zero replace merge strategy.
	o.Info = infos
	o.InfoChange = true
	o.HasInfo = true
	o.Changed = true
	w.changedObjects = append(w.changedObjects, o)
}

// GetObjects returns an array of all objects the client knows about.
func (w *World) GetObjects() []*Object {
	return w.objects
}

func (w *World) GetChangedObjects() []*Object {
	return w.changedObjects
}

func (w *World) ClearChangedObjects() {
	w.changedObjects = make([]*Object, 0)
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

func (w *World) getConeRays(originY, originX, originZ float64, radius float64, distance float64) (targets [][2][3]float64) {
	// Collect our end-points as 16 points aligned against the Y axis.
	/*points := 32
	rads := math.Pi * 2 / float64(points)
	for i := 0; i < points; i++ {
		x := math.Sin(rads*float64(i)) * radius
		y := math.Cos(rads*float64(i)) * radius
		z := distance
		targets = append(targets, [2][3]float64{
			{originY, originX, originZ},
			{originY + y, originX + x, originZ + z},
		})
	}*/
	r2 := radius * radius
	area := int(r2) << 2
	rr := int(radius) << 1

	for i := 0; i < area; i++ {
		tx := float64((i % rr) - int(radius))
		ty := float64((i / rr) - int(radius))

		if originY+ty < originY {
			continue
		}

		if tx*tx+ty*ty <= r2 {
			targets = append(targets, [2][3]float64{
				{originY, originX, originZ - radius/2},
				{originY + ty, originX + tx, originZ + radius},
			})
		}
	}

	return targets
}

func (w *World) getCubeRays(originY, originX, originZ float64, minY, minX, minZ, maxY, maxX, maxZ int) (c [][2][3]float64) {
	for y := minY; y < maxY; y++ {
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
	vhh := float64(w.viewHeight / 2)
	vwh := float64(w.viewWidth / 2)
	vdh := float64(w.viewDepth / 2)

	ymin := y1 - vhh
	if ymin < 0 {
		ymin = 0
	}
	ymax := y1 + vhh
	if ymax > float64(m.GetHeight()) {
		ymax = float64(m.GetHeight())
	}

	xmin := x1 - vwh
	if xmin < 0 {
		xmin = 0
	}
	xmax := x1 + vwh
	if xmax > float64(m.GetWidth()) {
		xmax = float64(m.GetWidth())
	}

	zmin := z1 - vdh
	if zmin < 0 {
		zmin = 0
	}
	zmax := z1 + vdh
	if zmax > float64(m.GetDepth()) {
		zmax = float64(m.GetDepth())
	}

	rays := w.getCubeRays(y1, x1, z1, int(ymin), int(xmin), int(zmin), int(ymax), int(xmax), int(zmax))

	visibleTiles := make([]bool, m.height*m.width*m.depth)

	markTiles := func(y, x, z int) bool {
		// NOTE: We're unsafely accessing map tiles since the results of our rayCasts are
		visibleTiles[m.height*m.width*z+m.width*y+x] = true
		return m.tiles[m.height*m.width*z+m.width*y+x].opaque
		/*tile := m.GetTile(y, x, z)
		if tile != nil {
			visibleTiles[y][x][z] = true

			if tile.opaque {
				return true
			}
		}
		return false*/
	}

	// This feels wrong, but we duplicate the rays and offset the origin to ensure we can see over vertical edges on character's sides.
	for _, r := range rays {
		rays = append(rays, [2][3]float64{
			{r[0][0], r[0][1] + float64(o.W), r[0][2] + float64(o.D)},
			{r[1][0], r[1][1], r[1][2]},
		})
	}

	// Now let's shoot some rays via Amanatides & Woo.
	w.rayCasts(rays, float64(m.GetHeight()), float64(m.GetWidth()), float64(m.GetDepth()), markTiles)
	// Set objects no longer visible
	for j := 0; j < len(m.tiles); j++ {
		isVisible := visibleTiles[j]
		for _, o := range m.tiles[j].objects {
			if !isVisible && o.Visible {
				o.Visible = false
				o.VisibilityChange = true
				w.changedObjects = append(w.changedObjects, o)
			} else if isVisible && !o.Visible {
				o.Visible = true
				o.VisibilityChange = true
				w.changedObjects = append(w.changedObjects, o)
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

	unblockedTiles := make([][][]bool, m.GetHeight())
	for i := range unblockedTiles {
		unblockedTiles[i] = make([][]bool, m.GetWidth())
		for j := range unblockedTiles[i] {
			unblockedTiles[i][j] = make([]bool, m.GetDepth())
		}
	}

	inRoom := func() bool {
		oY := float64(o.Y) + float64(o.H)/2
		oX := float64(o.X) + float64(o.W)/2
		oZ := float64(o.Z)

		// Now let's shoot some rays via Amanatides & Woo.
		hits := 0
		rays := [][2][3]float64{
			{{oY, oX, oZ}, {oY + 15, oX, oZ}},
			{{oY, oX, oZ}, {oY, oX - 25, oZ}},
			{{oY, oX, oZ}, {oY, oX + 25, oZ}},
			{{oY, oX, oZ}, {oY, oX, oZ - 25}},
			{{oY, oX, oZ}, {oY, oX, oZ + 25}},
		}
		w.rayCasts(rays, float64(m.GetHeight()), float64(m.GetWidth()), float64(m.GetDepth()), func(y, x, z int) bool {
			if m.tiles[m.height*m.width*z+m.width*y+x].opaque {
				hits++
				return true
			}
			return false
		})
		return hits >= 4
	}

	isAboveBlocked := func() bool {
		oY := float64(o.Y) + float64(o.H)/2
		oX := float64(o.X) + float64(o.W)/2
		oZ := float64(o.Z)

		// Now let's shoot some rays via Amanatides & Woo.
		hits := 0
		rays := [][2][3]float64{
			{{oY, oX, oZ}, {oY + 16, oX, oZ}},
		}
		w.rayCasts(rays, float64(m.GetHeight()), float64(m.GetWidth()), float64(m.GetDepth()), func(y, x, z int) bool {
			if m.tiles[m.height*m.width*z+m.width*y+x].opaque {
				hits++
				return true
			}
			return false
		})
		return hits >= 1
	}

	isLeftBlocked := func() bool {
		oY := float64(o.Y) + float64(o.H)/2
		oX := float64(o.X) + float64(o.W)/2
		oZ := float64(o.Z)

		// Now let's shoot some rays via Amanatides & Woo.
		hits := 0
		rays := [][2][3]float64{
			{{oY, oX, oZ}, {oY, oX - 2, oZ}},
			{{oY, oX, oZ}, {oY, oX - 2, oZ + 1}},
			{{oY, oX, oZ}, {oY, oX - 2, oZ + 2}},
			{{oY, oX, oZ}, {oY, oX - 2, oZ + 3}},
		}
		w.rayCasts(rays, float64(m.GetHeight()), float64(m.GetWidth()), float64(m.GetDepth()), func(y, x, z int) bool {
			if m.tiles[m.height*m.width*z+m.width*y+x].opaque {
				hits++
				return true
			}
			return false
		})
		return hits >= 2
	}

	isFrontBlocked := func() bool {
		oY := float64(o.Y) + float64(o.H)/2
		oX := float64(o.X) + float64(o.W)/2
		oZ := float64(o.Z)

		// Now let's shoot some rays via Amanatides & Woo.
		hits := 0
		rays := [][2][3]float64{
			{{oY, oX, oZ}, {oY, oX, oZ + 6}},
			{{oY, oX, oZ}, {oY, oX - 1, oZ + 8}},
			{{oY, oX, oZ}, {oY, oX - 2, oZ + 8}},
			{{oY, oX, oZ}, {oY, oX - 3, oZ + 8}},
			{{oY, oX, oZ}, {oY, oX + 1, oZ + 8}},
			{{oY, oX, oZ}, {oY, oX + 2, oZ + 8}},
			{{oY, oX, oZ}, {oY, oX + 3, oZ + 8}},
		}
		w.rayCasts(rays, float64(m.GetHeight()), float64(m.GetWidth()), float64(m.GetDepth()), func(y, x, z int) bool {
			if m.tiles[m.height*m.width*z+m.width*y+x].opaque {
				hits++
				return true
			}
			return false
		})
		return hits >= 3
	}

	w.InRoom = inRoom()
	w.LeftBlocked = isLeftBlocked()
	w.FrontBlocked = isFrontBlocked()
	w.AboveBlocked = isAboveBlocked()

	if w.InRoom {
		for y := o.Y + 1; y < o.Y+16; y++ {
			if y < 0 || y >= m.height {
				continue
			}
			for x := o.X - w.viewWidth/2; x < o.X+w.viewWidth/2; x++ {
				if x < 0 || x >= m.width {
					continue
				}
				for z := o.Z - w.viewDepth/2; z < o.Z+w.viewDepth/2; z++ {
					if z < 0 || z >= m.depth {
						continue
					}
					if m.tiles[m.height*m.width*z+m.width*y+x].opaque {
						unblockedTiles[y][x][z] = true
					}
				}
			}
		}
	} else if w.AboveBlocked || w.FrontBlocked {
		for y := o.Y; y < o.Y+16; y++ {
			if y < 0 || y >= m.height {
				continue
			}
			for x := o.X - w.viewWidth/2; x < o.X+w.viewWidth/2; x++ {
				if x < 0 || x >= m.width {
					continue
				}
				for z := o.Z; z < o.Z+w.viewDepth/2; z++ {
					if z < 0 || z >= m.depth {
						continue
					}
					if m.tiles[m.height*m.width*z+m.width*y+x].opaque {
						unblockedTiles[y][x][z] = true
					}
				}
			}
		}
	} else if w.LeftBlocked {
		for y := o.Y; y < o.Y+16; y++ {
			if y < 0 || y >= m.height {
				continue
			}
			for x := o.X - w.viewWidth/2; x < o.X; x++ {
				if x < 0 || x >= m.width {
					continue
				}
				for z := o.Z; z < o.Z+w.viewDepth/2; z++ {
					if z < 0 || z >= m.depth {
						continue
					}
					if m.tiles[m.height*m.width*z+m.width*y+x].opaque {
						unblockedTiles[y][x][z] = true
					}
				}
			}
		}
	}

	// Collect our end-points for rays
	/*oY := float64(o.Y) + float64(o.H)/2
	oX := float64(o.X) + float64(o.W)/2
	oZ := float64(o.Z)

	//rays := w.getConeRays(oY, oX, oZ, 16, 16)
	minY := int(oY) + 4
	maxY := minY + int(o.H) + 12
	minX := int(oX) - 14
	maxX := minX + int(o.W) + 20
	minZ := int(oZ)
	maxZ := minZ + int(o.D) + 20
	rays := w.getCubeRays(oY, oX, oZ, minY, minX, minZ, maxY, maxX, maxZ)
	// TODO: We actually need to use an angled cone, originating from the near view target origin to whatever area we deem as the "camera" area
	// TODO: Or, we could have 2 "cubes" -- basically 2 flat cubes that create a "right angle bracket"

	// Now let's shoot some rays via Amanatides & Woo.
	w.rayCasts(rays, float64(m.GetHeight()), float64(m.GetWidth()), float64(m.GetDepth()), func(y, x, z int) bool {
		if m.tiles[m.height*m.width*z+m.width*y+x].opaque {
			unblockedTiles[y][x][z] = true
		}
		return false
	})*/

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
						w.changedObjects = append(w.changedObjects, o)
					} else if isUnblocked && !o.Unblocked {
						o.Unblocked = true
						o.UnblockedChange = true
						w.changedObjects = append(w.changedObjects, o)
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
				} else {
					o.FrameIndex = 0
				}
				o.Animation = anim
				o.Face = face
				if o.FrameIndex < len(face.Frames) {
					o.Frame = &(face.Frames[o.FrameIndex])
				} else {
					fmt.Fprintf(os.Stderr, "missing frame %d\n", animationID)
				}
				if o.Frame != nil {
					img, err := w.dataManager.GetCachedImage(o.Frame.ImageID)
					if err != nil {
						w.PendingObjectImages[o.Frame.ImageID] = append(w.PendingObjectImages[o.Frame.ImageID], o.ID)
					} else {
						o.Image = img
					}
				}

				w.changedObjects = append(w.changedObjects, o)
				o.Changed = true
			}
		}
		delete(w.PendingObjectAnimations, animationID)
	}
}

func (w *World) CheckPendingObjectImageIDs(imageID uint32) {
	if pending, ok := w.PendingObjectImages[imageID]; ok {
		img, err := w.dataManager.GetCachedImage(imageID)
		if err == nil {
			for _, objectID := range pending {
				if o := w.GetObject(objectID); o != nil {
					w.changedObjects = append(w.changedObjects, o)
					o.Image = img
					o.Changed = true
				}
			}
			delete(w.PendingObjectImages, imageID)
		}
	}
}

// UpdateCubes updates the reach and intersect cubes.
func (w *World) updateCubes() {
	vo := w.viewObject
	if vo == nil {
		return
	}
	// Calculate reach cube.
	maxY := int(vo.H) + int(vo.Reach)*2
	maxX := int(vo.W) + int(vo.Reach)*2
	maxZ := int(vo.Reach) * 2
	if vo.D > 1 {
		maxZ += int(vo.D)
	}

	w.ReachCube = make([][][]struct{}, maxY)
	for y := range w.ReachCube {
		w.ReachCube[y] = make([][]struct{}, maxX)
		for x := range w.ReachCube[y] {
			w.ReachCube[y][x] = make([]struct{}, maxZ)
		}
	}

	// Calculate intersect cube.
	maxY = int(vo.H)
	maxX = int(vo.W)
	maxZ = int(vo.D)
	w.IntersectCube = make([][][]struct{}, maxY)
	for y := range w.IntersectCube {
		w.IntersectCube[y] = make([][]struct{}, maxX)
		for x := range w.IntersectCube[y] {
			w.IntersectCube[y][x] = make([]struct{}, maxZ)
		}
	}
}
