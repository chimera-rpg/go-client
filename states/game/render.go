package game

import (
	"fmt"
	"math"
	"time"

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
		renderX, renderY, _ := s.GetRenderPosition(m, o.Y, o.X, o.Z)
		scale := *s.objectsScale
		tileWidth := int(s.Client.AnimationsConfig.TileWidth)
		tileHeight := int(s.Client.AnimationsConfig.TileHeight)

		// Calculate object-specific offsets.
		offsetX := 0
		offsetY := 0
		if adjust, ok := s.Client.AnimationsConfig.Adjustments[cdata.ArchetypeType(o.Type)]; ok {
			offsetX += int(adjust.X)
			offsetY += int(adjust.Y)
		}

		x := float64(renderX) + float64(offsetX)*scale
		y := float64(renderY) + float64(offsetY)*scale

		// Adjust for centering based on target's sizing.
		x += (float64(int(o.W)*tileWidth) * scale) / 2
		y += (float64((int(o.H)*int(s.Client.AnimationsConfig.YStep.Y) + (int(o.H) * tileHeight))) * scale) / 2
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

	// Iterate over world messages.
	now := time.Now()
	for i := len(s.mapMessages) - 1; i >= 0; i-- {
		msg := s.mapMessages[i]
		if now.After(msg.destroyTime) {
			s.MapContainer.GetDisownChannel() <- msg.el
			msg.el.GetDestroyChannel() <- true
			s.mapMessages = append(s.mapMessages[:i], s.mapMessages[i+1:]...)
		} else {
			// TODO: Check if msg has associated object and if it has moved.
		}
	}

	return
}

// GetRenderPosition gets world to pixel coordinate positions for a given tile location.
func (s *Game) GetRenderPosition(m *world.DynamicMap, y, x, z uint32) (targetX, targetY, targetZ int) {
	scale := *s.objectsScale
	tileWidth := int(s.Client.AnimationsConfig.TileWidth)
	tileHeight := int(s.Client.AnimationsConfig.TileHeight)

	originX := 0
	originY := int(m.GetHeight()) * int(-s.Client.AnimationsConfig.YStep.Y)
	originX += int(y) * int(s.Client.AnimationsConfig.YStep.X)
	originY += int(y) * int(s.Client.AnimationsConfig.YStep.Y)
	originX += int(x) * tileWidth
	originY += int(z) * tileHeight

	indexZ := int(z)
	indexX := int(x)
	indexY := int(y)

	targetZ = (indexZ * int(m.GetHeight()) * int(m.GetWidth())) + (int(m.GetDepth()) * indexY) - (indexX)

	// Calculate our scaled pixel position at which to render.
	targetX = int(float64(originX)*scale) + 100
	targetY = int(float64(originY)*scale) + 100
	return
}

// RenderObject renders a given Object within a DynamicMap.
func (s *Game) RenderObject(o *world.Object, m *world.DynamicMap) {
	scale := *s.objectsScale
	tileWidth := int(s.Client.AnimationsConfig.TileWidth)
	tileHeight := int(s.Client.AnimationsConfig.TileHeight)
	// If the object is currently missing, hide it. FIXME: It'd be better to keep it on screen, but grayscale, if it is outside of player view. If in player view, then it should be hidden.
	if o != s.world.GetViewObject() {
		if t, ok := s.objectImages[o.ID]; ok {
			if o.Missing {
				t.GetUpdateChannel() <- ui.UpdateHidden(true)
				return
			}
			t.GetUpdateChannel() <- ui.UpdateHidden(false)
		}
	}
	frames := s.Client.DataManager.GetFace(o.AnimationID, o.FaceID)
	// Bail if there are no frames to render.
	if len(frames) == 0 {
		return
	}

	// Get our render position.
	x, y, zIndex := s.GetRenderPosition(m, o.Y, o.X, o.Z)
	zIndex += o.Index

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

	// Adjust our target position.
	x += int(float64(offsetX) * scale)
	y += int(float64(offsetY) * scale)

	// Calculate our scaled pixel position at which to render.
	w := int(float64(tileWidth) * scale)
	h := int(float64(tileHeight) * scale)

	img := s.Client.DataManager.GetCachedImage(frames[0].ImageID)
	if _, ok := s.objectImages[o.ID]; !ok {
		if img != nil {
			bounds := img.Bounds()
			w = int(float64(bounds.Max.X) * scale)
			h = int(float64(bounds.Max.Y) * scale)
			if (o.H > 1 || o.D > 1) && bounds.Max.Y > tileHeight {
				y -= h - int(float64(tileHeight)*scale)
			}
			s.objectImages[o.ID] = ui.NewImageElement(ui.ImageElementConfig{
				Style: fmt.Sprintf(`
							X %d
							Y %d
							W %d
							H %d
							ZIndex %d
						`, x, y, w, h, zIndex),
				Image:       img,
				PostOutline: true,
				Events: ui.Events{
					OnPressed: func(button uint8, x, y int32) bool {
						if button != 1 {
							return true
						}
						/*s.focusedImage.GetUpdateChannel() <- ui.UpdateDimensions{
							X: ui.Number{Value: float64(s.objectImages[o.ID].GetX())},
							Y: ui.Number{Value: float64(s.objectImages[o.ID].GetY())},
							W: ui.Number{Value: float64(s.objectImages[o.ID].GetWidth())},
							H: ui.Number{Value: float64(s.objectImages[o.ID].GetHeight())},
						}
						s.focusedImage.GetUpdateChannel() <- img*/
						if s.objectImages[o.ID].PixelHit(x, y) {
							s.inputChan <- FocusObject(o.ID)
							return false
						}
						return true
					},
				},
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
				Image:       img,
				PostOutline: true,
				Events: ui.Events{
					OnPressed: func(button uint8, x, y int32) bool {
						if button != 1 {
							return true
						}
						if s.objectImages[o.ID].PixelHit(x, y) {
							s.inputChan <- FocusObject(o.ID)
							return false
						}
						return true
					},
				},
			})
		}
		s.MapContainer.GetAdoptChannel() <- s.objectImages[o.ID]
	} else {
		if img != nil {
			if o.Changed {
				bounds := img.Bounds()
				w = int(float64(bounds.Max.X) * scale)
				h = int(float64(bounds.Max.Y) * scale)

				var sw, sh float64
				sw = float64(w)
				sh = float64(h)
				if o.Squeezing {
					sw = math.Max(float64(w-w/4), float64(tileWidth)*scale)
				}
				if o.Crouching {
					sh = math.Max(float64(h-h/3), float64(tileHeight)*scale)
				}

				if (o.H > 1 || o.D > 1) && bounds.Max.Y > tileHeight {
					y -= int(sh) - int(float64(tileHeight)*scale)
				}
				s.objectImages[o.ID].GetUpdateChannel() <- ui.UpdateDimensions{
					X: ui.Number{Value: float64(x)},
					Y: ui.Number{Value: float64(y)},
					W: ui.Number{Value: sw},
					H: ui.Number{Value: sh},
				}
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