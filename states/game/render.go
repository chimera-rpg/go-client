package game

import (
	"fmt"
	"math"
	"time"

	"github.com/chimera-rpg/go-client/data"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-client/world"
	cdata "github.com/chimera-rpg/go-common/data"
)

// HandleRender handles the rendering of our Game state.
func (s *Game) HandleRender(delta time.Duration) {
	var batchMessages = make([]ui.BatchMessage, 0, 1024)
	// FIXME: This is _very_ rough and is just for testing!
	m := s.world.GetCurrentMap()
	objects := s.world.GetObjects()

	viewObject := s.world.GetViewObject()
	if o := viewObject; o != nil {
		renderX, renderY, _ := s.GetRenderPosition(m, o.Y, o.X, o.Z)
		scale := *s.objectsScale
		tileWidth := int(s.Client.AnimationsConfig.TileWidth)
		tileHeight := int(s.Client.AnimationsConfig.TileHeight)

		// Calculate object-specific offsets.
		offsetX := 0
		offsetY := 0
		adjust, ok := s.Client.AnimationsConfig.GetAdjustment(cdata.ArchetypeType(o.Type))
		if ok {
			offsetX += int(adjust.X)
			offsetY += int(adjust.Y)
		}

		x := float64(renderX) + float64(offsetX)*scale
		y := float64(renderY) + float64(offsetY)*scale

		// Adjust for centering based on target's sizing.
		x += (float64(int(o.W)*tileWidth) * scale) / 2
		y += (float64((int(o.H)*s.Client.AnimationsConfig.YStep.Y + (int(o.H) * tileHeight))) * scale) / 2
		// Center within the map container.
		x -= float64(s.MapContainer.GetWidth()) / 2
		y -= float64(s.MapContainer.GetHeight()) / 2

		batchMessages = append(batchMessages, ui.BatchUpdateMessage{
			Target: &s.MapContainer,
			Update: ui.UpdateScroll{
				Left: ui.Number{Value: x},
				Top:  ui.Number{Value: y},
			},
		})
		/*s.MapContainer.GetUpdateChannel() <- ui.UpdateScroll{
			Left: ui.Number{Value: x},
			Top:  ui.Number{Value: y},
		}*/
	}

	// Iterate over world objects.
	for _, o := range objects {
		batchMessages = s.RenderObject(viewObject, o, m, delta, batchMessages)
	}

	// Iterate over world messages.
	now := time.Now()
	for i := len(s.mapMessages) - 1; i >= 0; i-- {
		msg := s.mapMessages[i]
		if !msg.destroyTime.Equal(time.Time{}) && now.After(msg.destroyTime) {
			batchMessages = append(batchMessages, ui.BatchDisownMessage{
				Parent: &s.MapContainer,
				Target: msg.el,
			})
			batchMessages = append(batchMessages, ui.BatchDestroyMessage{
				Target: msg.el,
			})
			//s.MapContainer.GetDisownChannel() <- msg.el
			//msg.el.GetDestroyChannel() <- true
			s.mapMessages = append(s.mapMessages[:i], s.mapMessages[i+1:]...)
		} else {
			// TODO: Check if msg has associated object and if it has moved.
			if msg.trackObject {
				o := s.world.GetObject(msg.objectID)
				if o != nil {
					x := o.X
					y := o.Y + uint32(o.H) + 1
					z := o.Z
					xPos, yPos, _ := s.GetRenderPosition(s.world.GetCurrentMap(), y, x, z)
					batchMessages = append(batchMessages, ui.BatchUpdateMessage{
						Target: msg.el,
						Update: ui.UpdateX{
							Number: ui.Number{Value: float64(xPos)},
						},
					})
					batchMessages = append(batchMessages, ui.BatchUpdateMessage{
						Target: msg.el,
						Update: ui.UpdateY{
							Number: ui.Number{Value: float64(yPos)},
						},
					})
					/*msg.el.GetUpdateChannel() <- ui.UpdateX{
						Number: ui.Number{Value: float64(xPos)},
					}
					msg.el.GetUpdateChannel() <- ui.UpdateY{
						Number: ui.Number{Value: float64(yPos)},
					}*/
				}
			}
			// Move message upwards if need be.
			if msg.floatY != 0 {
				batchMessages = append(batchMessages, ui.BatchUpdateMessage{
					Target: msg.el,
					Update: ui.UpdateY{
						Number: ui.Number{Value: msg.el.GetStyle().Y.Value + msg.floatY*float64(delta.Milliseconds())},
					},
				})
				/*msg.el.GetUpdateChannel() <- ui.UpdateY{
					Number: ui.Number{Value: msg.el.GetStyle().Y.Value + msg.floatY*float64(delta.Milliseconds())},
				}*/
			}
		}
	}

	if len(batchMessages) > 10 {
		fmt.Println("batch", len(batchMessages))
		alphaCount := 0
		dimensionsCount := 0
		hiddenCount := 0
		grayCount := 0
		imageCount := 0
		otherCount := 0
		for _, m := range batchMessages {
			if m, ok := m.(ui.BatchUpdateMessage); ok {
				switch m.Update.(type) {
				case ui.UpdateAlpha:
					alphaCount++
				case ui.UpdateHidden:
					hiddenCount++
				case ui.UpdateGrayscale:
					grayCount++
				case ui.UpdateImageID:
					imageCount++
				case ui.UpdateDimensions:
					dimensionsCount++
				default:
					otherCount++
				}
			}
		}
		fmt.Printf("%d alpha, %d dim, %d hid, %d gray, %d img, %d other\n", alphaCount, dimensionsCount, hiddenCount, grayCount, imageCount, otherCount)
	}

	s.Client.RootWindow.BatchChannel <- batchMessages
	return
}

// GetRenderPosition gets world to pixel coordinate positions for a given tile location.
func (s *Game) GetRenderPosition(m *world.DynamicMap, y, x, z uint32) (targetX, targetY, targetZ int) {
	scale := *s.objectsScale
	tileWidth := int(s.Client.AnimationsConfig.TileWidth)
	tileHeight := int(s.Client.AnimationsConfig.TileHeight)

	originX := 0
	originY := int(m.GetHeight()) * -s.Client.AnimationsConfig.YStep.Y
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
func (s *Game) RenderObject(viewObject *world.Object, o *world.Object, m *world.DynamicMap, dt time.Duration, uiMessages []ui.BatchMessage) []ui.BatchMessage {
	if o != viewObject {
		if o.Element != nil {
			if o.Missing && o.WasMissing {
				return uiMessages
			}

			if o.Missing && !o.WasMissing {
				o.WasMissing = true
				uiMessages = append(uiMessages, ui.BatchUpdateMessage{
					Target: o.Element,
					Update: ui.UpdateHidden(true),
				})
				return uiMessages
			}
			if !o.Missing && o.WasMissing {
				o.WasMissing = false
				uiMessages = append(uiMessages, ui.BatchUpdateMessage{
					Target: o.Element,
					Update: ui.UpdateHidden(false),
				})
			}
		}
	}
	// Bail if there are no frames yet. FIXME: This should still show _something_
	if len(o.Face.Frames) == 0 {
		return uiMessages
	}
	o.Process(dt)

	frame := o.Face.Frames[o.FrameIndex]

	// Get and cache our render position.
	if o.Changed {
		o.RenderX, o.RenderY, o.RenderZ = s.GetRenderPosition(m, o.Y, o.X, o.Z)
		o.RenderZ += o.Index
		o.RecalculateFinalRender = true
	}
	x := o.RenderX
	y := o.RenderY
	zIndex := o.RenderZ

	// Acquire and cache our object's adjustment.
	if !o.Adjusted {
		adjust, _ := s.Client.AnimationsConfig.GetAdjustment(cdata.ArchetypeType(o.Type))
		o.AdjustX = int(adjust.X)
		o.AdjustY = int(adjust.Y)
		o.Adjusted = true
		o.RecalculateFinalRender = true
	}

	// Calculate archetype type-specific offsets.
	var offsetX, offsetY int
	var w, h int
	if o.RecalculateFinalRender {
		o.RecalculateFinalRender = false

		scale := *s.objectsScale
		tileWidth := s.Client.AnimationsConfig.TileWidth
		tileHeight := s.Client.AnimationsConfig.TileHeight

		offsetX += o.AdjustX
		offsetY += o.AdjustY

		// Set animation frame offsets.
		offsetX += int(frame.X)
		offsetY += int(frame.Y)

		// Adjust our target position.
		x += int(float64(offsetX) * scale)
		y += int(float64(offsetY) * scale)

		// Calculate our scaled pixel position at which to render.
		w = int(float64(tileWidth) * scale)
		h = int(float64(tileHeight) * scale)

		o.FinalRenderOffsetX = offsetX
		o.FinalRenderOffsetY = offsetY
		o.FinalRenderX = x
		o.FinalRenderY = y
		o.FinalRenderW = w
		o.FinalRenderH = h
	}
	x = o.FinalRenderX
	y = o.FinalRenderY
	offsetX = o.FinalRenderOffsetX
	offsetY = o.FinalRenderOffsetY
	w = o.FinalRenderW
	h = o.FinalRenderH

	// Get/create our shadow position, if we should.
	if o.HasShadow {
		uiMessages = s.RenderObjectShadows(o, m, offsetX, offsetY, w, h, uiMessages)
	}

	uiMessages = s.RenderObjectImage(o, m, frame, x, y, zIndex, w, h, uiMessages)
	return uiMessages
}

func (s *Game) RenderObjectImage(o *world.Object, m *world.DynamicMap, frame data.AnimationFrame, x, y, zIndex, w, h int, uiMessages []ui.BatchMessage) []ui.BatchMessage {
	img := o.Image
	if img == nil {
		var err error
		img, err = s.Client.DataManager.GetCachedImage(frame.ImageID)
		if err == nil {
			o.Image = img
		}
	}

	scale := *s.objectsScale
	tileWidth := s.Client.AnimationsConfig.TileWidth
	tileHeight := s.Client.AnimationsConfig.TileHeight

	if o.Element == nil {
		if img != nil {
			bounds := img.Bounds()
			w = int(float64(bounds.Max.X) * scale)
			h = int(float64(bounds.Max.Y) * scale)
			if (o.H > 1 || o.D > 1) && bounds.Max.Y > tileHeight {
				y -= h - int(float64(tileHeight)*scale)
			}
			o.Element = ui.NewImageElement(ui.ImageElementConfig{
				Style: fmt.Sprintf(`
							X %d
							Y %d
							W %d
							H %d
							ZIndex %d
						`, x, y, w, h, zIndex),
				ImageID:     frame.ImageID,
				PostOutline: true,
				Events: ui.Events{
					OnPressed: func(button uint8, x, y int32) bool {
						if button != 1 {
							return true
						}
						if o.Element.PixelHit(x, y) {
							s.inputChan <- FocusObject(o.ID)
							return false
						}
						return true
					},
				},
			})
		} else {
			o.Element = ui.NewImageElement(ui.ImageElementConfig{
				Style: fmt.Sprintf(`
							X %d
							Y %d
							W %d
							H %d
							ZIndex %d
						`, x, y, w, h, zIndex),
				ImageID:     frame.ImageID,
				PostOutline: true,
				Events: ui.Events{
					OnPressed: func(button uint8, x, y int32) bool {
						if button != 1 {
							return true
						}
						if o.Element.PixelHit(x, y) {
							s.inputChan <- FocusObject(o.ID)
							return false
						}
						return true
					},
				},
			})
		}
		uiMessages = append(uiMessages, ui.BatchAdoptMessage{
			Parent: &s.MapContainer,
			Target: o.Element,
		})
	} else {
		if img != nil {
			if o.UnblockedChange {
				if o.Unblocked {
					uiMessages = append(uiMessages, ui.BatchUpdateMessage{
						Target: o.Element,
						Update: ui.UpdateAlpha(0.2),
					})
				} else {
					uiMessages = append(uiMessages, ui.BatchUpdateMessage{
						Target: o.Element,
						Update: ui.UpdateAlpha(1.0),
					})
				}
				o.UnblockedChange = false
			}
			if o.VisibilityChange {
				if o.Visible {
					uiMessages = append(uiMessages, ui.BatchUpdateMessage{
						Target: o.Element,
						Update: ui.UpdateGrayscale(false),
					})
				} else {
					uiMessages = append(uiMessages, ui.BatchUpdateMessage{
						Target: o.Element,
						Update: ui.UpdateGrayscale(true),
					})
				}
				o.VisibilityChange = false
			}
			if o.LightingChange {
				uiMessages = append(uiMessages, ui.BatchUpdateMessage{
					Target: o.Element,
					Update: ui.UpdateColorMod{
						R: uint8(255 * o.Brightness),
						G: uint8(255 * o.Brightness),
						B: uint8(255 * o.Brightness),
						A: 255},
				})
				o.LightingChange = false
			}
			//
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

				uiMessages = append(uiMessages, ui.BatchUpdateMessage{
					Target: o.Element,
					Update: ui.UpdateDimensions{
						X: ui.Number{Value: float64(x)},
						Y: ui.Number{Value: float64(y)},
						W: ui.Number{Value: sw},
						H: ui.Number{Value: sh},
					},
				})
				uiMessages = append(uiMessages, ui.BatchUpdateMessage{
					Target: o.Element,
					Update: ui.UpdateZIndex{Number: ui.Number{Value: float64(zIndex)}},
				})
				o.Changed = false
			}
			// Only update the image if the image ID has changed.
			if o.FrameImageID != frame.ImageID {
				o.FrameImageID = frame.ImageID
				uiMessages = append(uiMessages, ui.BatchUpdateMessage{
					Target: o.Element,
					Update: ui.UpdateImageID(o.FrameImageID),
				})
			}
		}
	}
	return uiMessages
}

func (s *Game) RenderObjectShadows(o *world.Object, m *world.DynamicMap, offsetX, offsetY, w, h int, uiMessages []ui.BatchMessage) []ui.BatchMessage {
	scale := *s.objectsScale
	// TODO: We should probably slice up an object's shadows based upon its width and depth. This will probably require using polygons unless SDL_gfx can clip rendered ellipses. Or, perhaps, use SDL_gfx's pie drawing calls for each shadow quadrant?
	sy, sx, sz := s.world.GetObjectShadowPosition(o)

	x, y, zIndex := s.GetRenderPosition(m, uint32(sy), uint32(sx), uint32(sz))
	// TODO: Fix shadows so they have a higher zIndex than z+1, but only for y of the same.
	zIndex--

	// Adjust our target position.
	x += int(float64(offsetX) * scale)
	y += int((float64(offsetY) + float64(o.D)) * scale)

	w = w * int(o.W)
	h = h * int(o.D)

	// Reduce shadow by 1/4th if it is an item
	if o.Type == cdata.ArchetypeItem.AsUint8() {

		rw := w / 4
		rh := h / 4

		w -= rw
		h -= rh

		x += rw / 2
		y += rh / 2
	}

	if o.ShadowElement == nil {
		o.ShadowElement = ui.NewPrimitiveElement(ui.PrimitiveElementConfig{
			Shape: ui.EllipseShape,
			Style: fmt.Sprintf(`
							X %d
							Y %d
							W %d
							H %d
							ZIndex %d
							BackgroundColor 0 0 0 96
						`, x, y, w, h, zIndex),
		})
		//s.MapContainer.GetAdoptChannel() <- s.objectShadows[o.ID]
		uiMessages = append(uiMessages, ui.BatchAdoptMessage{
			Parent: &s.MapContainer,
			Target: o.ShadowElement,
		})
	} else {
		if o.Changed {
			/*s.objectShadows[o.ID].GetUpdateChannel() <- ui.UpdateDimensions{
				X: ui.Number{Value: float64(x)},
				Y: ui.Number{Value: float64(y)},
				W: ui.Number{Value: float64(w)},
				H: ui.Number{Value: float64(h)},
			}
			s.objectShadows[o.ID].GetUpdateChannel() <- ui.UpdateZIndex{Number: ui.Number{Value: float64(zIndex)}}*/

			uiMessages = append(uiMessages, ui.BatchUpdateMessage{
				Target: o.ShadowElement,
				Update: ui.UpdateDimensions{
					X: ui.Number{Value: float64(x)},
					Y: ui.Number{Value: float64(y)},
					W: ui.Number{Value: float64(w)},
					H: ui.Number{Value: float64(h)},
				},
			})

			uiMessages = append(uiMessages, ui.BatchUpdateMessage{
				Target: o.ShadowElement,
				Update: ui.UpdateZIndex{Number: ui.Number{Value: float64(zIndex)}},
			})

		}
	}
	return uiMessages
}
