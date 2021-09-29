package states

import (
	"fmt"

	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-client/world"
	cdata "github.com/chimera-rpg/go-common/data"
)

// HandleRender handles the rendering of our Game state.
func (s *Game) HandleRender() {
	// FIXME: This is _very_ rough and is just for testing!
	m := s.world.GetCurrentMap()
	objects := s.world.GetObjects()
	// Delete images that no longer correspond to an existing world object.
	for oID, t := range s.objectImages {
		o := s.world.GetObject(oID)
		if o == nil {
			t.GetDestroyChannel() <- true
			delete(s.objectImages, oID)
		}
	}

	if o := s.world.GetViewObject(); o != nil {
		scale := 4
		tileWidth := int(s.Client.AnimationsConfig.TileWidth)
		tileHeight := int(s.Client.AnimationsConfig.TileHeight)

		originX := 0
		originY := int(m.GetHeight()) * int(-s.Client.AnimationsConfig.YStep.Y)
		originX += int(o.Y) * int(s.Client.AnimationsConfig.YStep.X)
		originY += int(o.Y) * int(s.Client.AnimationsConfig.YStep.Y)
		originX += int(o.X) * tileWidth
		originY += int(o.Z) * tileHeight
		// Calculate object-specific offsets.
		offsetX := 0
		offsetY := 0
		if adjust, ok := s.Client.AnimationsConfig.Adjustments[cdata.ArchetypeType(o.Type)]; ok {
			offsetX += int(adjust.X)
			offsetY += int(adjust.Y)
		}

		// Calculate our scaled pixel position at which to render.
		x := float64((originX+offsetX)*scale + 100)
		y := float64((originY+offsetY)*scale + 100)
		// Adjust for centering based on target's sizing.
		x += float64(int(o.W)*tileWidth*scale) / 2
		y += float64((int(o.H)*int(s.Client.AnimationsConfig.YStep.Y)+(int(o.H)*tileHeight))*scale) / 2
		// Center within the map container.
		x -= float64(s.MapContainer.GetWidth()) / 2
		y -= float64(s.MapContainer.GetHeight()) / 2

		s.MapContainer.GetUpdateChannel() <- ui.UpdateScroll{
			Left: ui.Number{Value: x},
			Top:  ui.Number{Value: y},
		}
	}

	// Iterate over world objects.
	for _, o := range objects {
		s.RenderObject(o, m)
	}
	return
}

// RenderObject renders a given Object within a DynamicMap.
func (s *Game) RenderObject(o *world.Object, m *world.DynamicMap) {
	scale := 4
	tileWidth := int(s.Client.AnimationsConfig.TileWidth)
	tileHeight := int(s.Client.AnimationsConfig.TileHeight)
	// If the object is missing (out of view), delete it. FIXME: This should probably convert the image rendering to semi-opaque or otherwise instead.
	if o.Missing {
		if t, ok := s.objectImages[o.ID]; ok {
			t.GetDestroyChannel() <- true
			delete(s.objectImages, o.ID)
			delete(s.objectImageIDs, o.ID)
		}
		return
	}
	frames := s.Client.DataManager.GetFace(o.AnimationID, o.FaceID)
	// Bail if there are no frames to render.
	if len(frames) == 0 {
		return
	}
	// Calculate our origin.
	originX := 0
	originY := int(m.GetHeight()) * int(-s.Client.AnimationsConfig.YStep.Y)
	originX += int(o.Y) * int(s.Client.AnimationsConfig.YStep.X)
	originY += int(o.Y) * int(s.Client.AnimationsConfig.YStep.Y)
	originX += int(o.X) * tileWidth
	originY += int(o.Z) * tileHeight

	// Calculate archetype type-specific offsets.
	offsetX := 0
	offsetY := 0
	if adjust, ok := s.Client.AnimationsConfig.Adjustments[cdata.ArchetypeType(o.Type)]; ok {
		offsetX += int(adjust.X)
		offsetY += int(adjust.Y)
	}

	// Set animation frame offsets.
	offsetX += int(frames[0].X)
	offsetY += int(frames[0].Y)

	// Get our render z-index.
	indexZ := int(o.Z)
	indexX := int(o.X)
	indexY := int(o.Y)

	zIndex := (indexZ * int(m.GetHeight()) * int(m.GetWidth())) + (int(m.GetDepth()) * indexY) - (indexX) + o.Index

	// Calculate our scaled pixel position at which to render.
	x := (originX+offsetX)*scale + 100
	y := (originY+offsetY)*scale + 100
	w := tileWidth * scale
	h := tileHeight * scale

	img := s.Client.DataManager.GetCachedImage(frames[0].ImageID)
	if _, ok := s.objectImages[o.ID]; !ok {
		if img != nil {
			bounds := img.Bounds()
			w = bounds.Max.X * scale
			h = bounds.Max.Y * scale
			if (o.H > 1 || o.D > 1) && bounds.Max.Y > tileHeight {
				y -= h - (tileHeight * scale)
			}
			s.objectImages[o.ID] = ui.NewImageElement(ui.ImageElementConfig{
				Style: fmt.Sprintf(`
							X %d
							Y %d
							W %d
							H %d
							ZIndex %d
						`, x, y, w, h, zIndex),
				Image: img,
			})
			s.objectImageIDs[o.ID] = frames[0].ImageID
		} else {
			s.objectImages[o.ID] = ui.NewImageElement(ui.ImageElementConfig{
				Style: fmt.Sprintf(`
							X %d
							Y %d
							W %d
							H %d
							ZIndex %d
						`, x, y, w, h, zIndex),
				Image: img,
			})
		}
		s.MapContainer.GetAdoptChannel() <- s.objectImages[o.ID]
	} else {
		if img != nil {
			bounds := img.Bounds()
			w = bounds.Max.X * scale
			h = bounds.Max.Y * scale
			if (o.H > 1 || o.D > 1) && bounds.Max.Y > tileHeight {
				y -= h - (tileHeight * scale)
			}
			if o.Changed {
				s.objectImages[o.ID].GetUpdateChannel() <- ui.UpdateX{Number: ui.Number{Value: float64(x)}}
				s.objectImages[o.ID].GetUpdateChannel() <- ui.UpdateY{Number: ui.Number{Value: float64(y)}}
				s.objectImages[o.ID].GetUpdateChannel() <- ui.UpdateW{Number: ui.Number{Value: float64(w)}}
				s.objectImages[o.ID].GetUpdateChannel() <- ui.UpdateH{Number: ui.Number{Value: float64(h)}}
				s.objectImages[o.ID].GetUpdateChannel() <- ui.UpdateZIndex{Number: ui.Number{Value: float64(zIndex)}}
				o.Changed = false
			}
			// Only update the image if the image ID has changed.
			if s.objectImageIDs[o.ID] != frames[0].ImageID {
				s.objectImageIDs[o.ID] = frames[0].ImageID
				s.objectImages[o.ID].GetUpdateChannel() <- img
			}
		}
	}
}
