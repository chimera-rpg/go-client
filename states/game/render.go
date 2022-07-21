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
	// Delete images that no longer correspond to an existing world object.
	/*temp := s.objectImages[:0]
	for _, or := range s.objectImages {
		o := s.world.GetObject(or.ID)
		if o == nil {
			or.el.GetDestroyChannel() <- true
		} else {
			temp = append(temp, or)
		}
	}
	s.objectImages = temp*/
	/*for oID, t := range s.objectImages {
		o := s.world.GetObject(oID)
		if o == nil {
			t.GetDestroyChannel() <- true
			delete(s.objectImages, oID)
		}
	}*/
	for oID, t := range s.objectShadows {
		o := s.world.GetObject(oID)
		if o == nil {
			fmt.Println("Destroying shadow for ", oID)
			//t.GetDestroyChannel() <- true
			batchMessages = append(batchMessages, ui.BatchDestroyMessage{
				Target: t,
			})
			delete(s.objectShadows, oID)
		}
	}

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
		y += (float64((int(o.H)*int(s.Client.AnimationsConfig.YStep.Y) + (int(o.H) * tileHeight))) * scale) / 2
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
func (s *Game) RenderObject(viewObject *world.Object, o *world.Object, m *world.DynamicMap, dt time.Duration, uiMessages []ui.BatchMessage) []ui.BatchMessage {
	scale := *s.objectsScale
	tileWidth := int(s.Client.AnimationsConfig.TileWidth)
	tileHeight := int(s.Client.AnimationsConfig.TileHeight)

	if o != viewObject {
		if o.Element != nil {
			/*if o.OutOfVisionChanged {
				o.OutOfVisionChanged = false
				if o.OutOfVision {
					uiMessages = append(uiMessages, ui.BatchUpdateMessage{
						Target: o.Element,
						Update: ui.UpdateHidden(true),
					})
					return uiMessages
				} else {
					uiMessages = append(uiMessages, ui.BatchUpdateMessage{
						Target: o.Element,
						Update: ui.UpdateHidden(false),
					})
				}
			}

			if o.OutOfVision {
				return uiMessages
			}*/

			if o.Missing && o.WasMissing {
				return uiMessages
			}

			if o.Missing && !o.WasMissing {
				o.WasMissing = true
				//o.Element.GetUpdateChannel() <- ui.UpdateHidden(true)
				uiMessages = append(uiMessages, ui.BatchUpdateMessage{
					Target: o.Element,
					Update: ui.UpdateHidden(true),
				})
				return uiMessages
			}
			if !o.Missing && o.WasMissing {
				o.WasMissing = false
				//o.Element.GetUpdateChannel() <- ui.UpdateHidden(false)
				uiMessages = append(uiMessages, ui.BatchUpdateMessage{
					Target: o.Element,
					Update: ui.UpdateHidden(false),
				})
			}
		}
	}
	// Bail if there is no animation yet. FIXME: This should still show _something_
	if o.Animation == nil {
		return uiMessages
	}
	frames := o.Animation.GetFace(o.FaceID)
	// Bail if there are no frames to render.
	if len(frames) == 0 {
		return uiMessages
	}
	// Check for frameindex oob, as the animation or face might have changed.
	if o.FrameIndex >= len(frames) {
		o.FrameIndex = len(frames) - 1
	}
	frame := frames[o.FrameIndex]

	// Animate if there are frames and they are visible. NOTE: We *might* want to be able to flag particular animations as requiring having their frames constantly elapsed, or simply record the current real frame and only update the corresponding image render when visibility is restored.
	if len(frames) > 1 && frame.Time > 0 && o.Visible {
		o.FrameElapsed += dt
		for ft := time.Duration(frame.Time) * time.Millisecond; o.FrameElapsed >= ft; {
			o.FrameElapsed -= ft
			o.FrameIndex++
			if o.FrameIndex >= len(frames) {
				o.FrameIndex = 0
			}
			frame = frames[o.FrameIndex]
			ft = time.Duration(frame.Time) * time.Millisecond
		}
	}

	// Get our render position.
	x, y, zIndex := s.GetRenderPosition(m, o.Y, o.X, o.Z)
	zIndex += o.Index

	// Calculate archetype type-specific offsets.
	offsetX := 0
	offsetY := 0
	adjust, ok := s.Client.AnimationsConfig.GetAdjustment(cdata.ArchetypeType(o.Type))
	if ok {
		offsetX += int(adjust.X)
		offsetY += int(adjust.Y)
	}

	// Set animation frame offsets.
	offsetX += int(frame.X)
	offsetY += int(frame.Y)

	// Adjust our target position.
	x += int(float64(offsetX) * scale)
	y += int(float64(offsetY) * scale)

	// Calculate our scaled pixel position at which to render.
	w := int(float64(tileWidth) * scale)
	h := int(float64(tileHeight) * scale)

	// Get/create our shadow position, if we should.
	if o.Type == cdata.ArchetypeNPC.AsUint8() || o.Type == cdata.ArchetypePC.AsUint8() || o.Type == cdata.ArchetypeItem.AsUint8() {
		uiMessages = s.RenderObjectShadows(o, m, offsetX, offsetY, w, h, uiMessages)
	}

	uiMessages = s.RenderObjectImage(o, m, frame, x, y, zIndex, w, h, uiMessages)
	return uiMessages
}

func (s *Game) RenderObjectImage(o *world.Object, m *world.DynamicMap, frame data.AnimationFrame, x, y, zIndex, w, h int, uiMessages []ui.BatchMessage) []ui.BatchMessage {
	scale := *s.objectsScale
	tileWidth := int(s.Client.AnimationsConfig.TileWidth)
	tileHeight := int(s.Client.AnimationsConfig.TileHeight)

	img := o.Image
	if img == nil {
		var err error
		img, err = s.Client.DataManager.GetCachedImage(frame.ImageID)
		if err == nil {
			o.Image = img
		}
	}
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
				//Image:       img,
				ImageID:     frame.ImageID,
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
				//Image:       img,
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
			//s.addObjectImage(o.ID, o.Element)
		}
		//s.MapContainer.GetAdoptChannel() <- o.Element
		uiMessages = append(uiMessages, ui.BatchAdoptMessage{
			Parent: &s.MapContainer,
			Target: o.Element,
		})
	} else {
		if img != nil {
			if o.UnblockedChange {
				if o.Unblocked {
					//o.Element.GetUpdateChannel() <- ui.UpdateAlpha(0.2)
					uiMessages = append(uiMessages, ui.BatchUpdateMessage{
						Target: o.Element,
						Update: ui.UpdateAlpha(0.2),
					})
				} else {
					//o.Element.GetUpdateChannel() <- ui.UpdateAlpha(1.0)
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
					//o.Element.GetUpdateChannel() <- ui.UpdateGrayscale(false)
				} else {
					//o.Element.GetUpdateChannel() <- ui.UpdateGrayscale(true)
					uiMessages = append(uiMessages, ui.BatchUpdateMessage{
						Target: o.Element,
						Update: ui.UpdateGrayscale(true),
					})
				}
				o.VisibilityChange = false
			}
			if o.LightingChange {
				/*o.Element.GetUpdateChannel() <- ui.UpdateColorMod{
				R: uint8(255 * o.Brightness),
				G: uint8(255 * o.Brightness),
				B: uint8(255 * o.Brightness),
				A: 255}*/
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
				/*o.Element.GetUpdateChannel() <- ui.UpdateDimensions{
					X: ui.Number{Value: float64(x)},
					Y: ui.Number{Value: float64(y)},
					W: ui.Number{Value: sw},
					H: ui.Number{Value: sh},
				}*/
				uiMessages = append(uiMessages, ui.BatchUpdateMessage{
					Target: o.Element,
					Update: ui.UpdateZIndex{Number: ui.Number{Value: float64(zIndex)}},
				})
				//o.Element.GetUpdateChannel() <- ui.UpdateZIndex{Number: ui.Number{Value: float64(zIndex)}}
				o.Changed = false
			}
			// Only update the image if the image ID has changed.
			if o.FrameImageID != frame.ImageID {
				o.FrameImageID = frame.ImageID
				//o.Element.GetUpdateChannel() <- img
				uiMessages = append(uiMessages, ui.BatchUpdateMessage{
					Target: o.Element,
					Update: ui.UpdateImageID(o.FrameImageID),
				})
				//o.Element.GetUpdateChannel() <- ui.UpdateImageID(o.FrameImageID)
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

	if _, ok := s.objectShadows[o.ID]; !ok {
		s.objectShadows[o.ID] = ui.NewPrimitiveElement(ui.PrimitiveElementConfig{
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
			Target: s.objectShadows[o.ID],
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
				Target: s.objectShadows[o.ID],
				Update: ui.UpdateDimensions{
					X: ui.Number{Value: float64(x)},
					Y: ui.Number{Value: float64(y)},
					W: ui.Number{Value: float64(w)},
					H: ui.Number{Value: float64(h)},
				},
			})

			uiMessages = append(uiMessages, ui.BatchUpdateMessage{
				Target: s.objectShadows[o.ID],
				Update: ui.UpdateZIndex{Number: ui.Number{Value: float64(zIndex)}},
			})

		}
	}
	return uiMessages
}
