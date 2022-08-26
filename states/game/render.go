package game

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/chimera-rpg/go-client/data"
	"github.com/chimera-rpg/go-client/states/game/elements"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-client/world"
	cdata "github.com/chimera-rpg/go-common/data"
)

type RenderContext struct {
	scale                             float64
	tileWidth, tileHeight             int
	tileWidthScaled, tileHeightScaled int
}

func (s *Game) GetRenderContext() RenderContext {
	return RenderContext{
		scale:            *s.objectsScale,
		tileWidth:        s.Client.AnimationsConfig.TileWidth,
		tileHeight:       s.Client.AnimationsConfig.TileHeight,
		tileWidthScaled:  int(float64(s.Client.AnimationsConfig.TileWidth) * *s.objectsScale),
		tileHeightScaled: int(float64(s.Client.AnimationsConfig.TileHeight) * *s.objectsScale),
	}
}

type BatchMessages struct {
	messages []ui.BatchMessage
}

func (b *BatchMessages) add(m ui.BatchMessage) {
	b.messages = append(b.messages, m)
}

// HandleRender handles the rendering of our Game state.
func (s *Game) HandleRender(delta time.Duration) {
	var batchMessages BatchMessages
	ctx := s.GetRenderContext()
	// FIXME: This is _very_ rough and is just for testing!
	m := s.world.GetCurrentMap()

	viewObject := s.world.GetViewObject()
	if o := viewObject; o != nil && o.Changed {
		renderX, renderY, _ := s.GetRenderPosition(ctx, m, o.Y, o.X, o.Z)

		// Calculate object-specific offsets.
		offsetX := 0
		offsetY := 0
		adjust, ok := s.Client.AnimationsConfig.GetAdjustment(cdata.ArchetypeType(o.Type))
		if ok {
			offsetX += int(adjust.X)
			offsetY += int(adjust.Y)
		}

		x := float64(renderX) + float64(offsetX)*ctx.scale
		y := float64(renderY) + float64(offsetY)*ctx.scale

		// Adjust for centering based on target's sizing.
		x += (float64(int(o.W)*ctx.tileWidth) * ctx.scale) / 2
		y += (float64((int(o.H)*s.Client.AnimationsConfig.YStep.Y + (int(o.H) * ctx.tileHeight))) * ctx.scale) / 2
		// Center within the map container.
		x -= float64(s.MapWindow.Container.GetWidth()) / 2
		y -= float64(s.MapWindow.Container.GetHeight()) / 2

		batchMessages.add(ui.BatchUpdateMessage{
			Target: &s.MapWindow.Container,
			Update: ui.UpdateScroll{
				Left: ui.Number{Value: x},
				Top:  ui.Number{Value: y},
			},
		})

		s.RenderObject(ctx, viewObject, o, m, delta, &batchMessages)
	}

	// Iterate over world objects.
	objects := s.world.GetChangedObjects()
	if len(objects) > 1000 {
		fmt.Printf("Rendering %d objects\n", len(objects))
	}
	for _, o := range objects {
		s.RenderObject(ctx, viewObject, o, m, delta, &batchMessages)
	}

	// FIXME: This was moved from the viewObject check so as to allow updating beyond when the view object has changed.
	if viewObject != nil {
		// FIXME: We should keep track of tile mod time, then tell our ground window to refresh its tiles if any of those tiles have changed.
		if len(objects) > 0 {
			s.GroundWindow.Refresh()
			s.InspectorWindow.Refresh()
		}
	}

	s.world.ClearChangedObjects()

	// Iterate over world messages.
	now := time.Now()
	for i := len(s.MapWindow.Messages) - 1; i >= 0; i-- {
		msg := s.MapWindow.Messages[i]
		if !msg.DestroyTime.Equal(time.Time{}) && now.After(msg.DestroyTime) {
			batchMessages.add(ui.BatchDisownMessage{
				Parent: &s.MapWindow.Container,
				Target: msg.El,
			})
			batchMessages.add(ui.BatchDestroyMessage{
				Target: msg.El,
			})
			s.MapWindow.Messages = append(s.MapWindow.Messages[:i], s.MapWindow.Messages[i+1:]...)
		} else {
			// TODO: Check if msg has associated object and if it has moved.
			if msg.TrackObject {
				o := s.world.GetObject(msg.ObjectID)
				if o != nil {
					x := o.X
					y := o.Y + int(o.H) + 1
					z := o.Z
					xPos, yPos, _ := s.GetRenderPosition(ctx, s.world.GetCurrentMap(), y, x, z)
					batchMessages.add(ui.BatchUpdateMessage{
						Target: msg.El,
						Update: ui.UpdateX{
							Number: ui.Number{Value: float64(xPos)},
						},
					})
					batchMessages.add(ui.BatchUpdateMessage{
						Target: msg.El,
						Update: ui.UpdateY{
							Number: ui.Number{Value: float64(yPos)},
						},
					})
				}
			}
			// Move message upwards if need be.
			if msg.FloatY != 0 {
				batchMessages.add(ui.BatchUpdateMessage{
					Target: msg.El,
					Update: ui.UpdateY{
						Number: ui.Number{Value: msg.El.GetStyle().Y.Value + msg.FloatY*float64(delta.Milliseconds())},
					},
				})
			}
		}
	}

	if len(batchMessages.messages) > 10 {
		fmt.Println("batch", len(batchMessages.messages))
		alphaCount := 0
		dimensionsCount := 0
		hiddenCount := 0
		grayCount := 0
		imageCount := 0
		otherCount := 0
		for _, m := range batchMessages.messages {
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

	s.Client.RootWindow.BatchChannel <- batchMessages.messages
	return
}

// GetRenderPosition gets world to pixel coordinate positions for a given tile location.
func (s *Game) GetRenderPosition(ctx RenderContext, m *world.DynamicMap, y, x, z int) (targetX, targetY, targetZ int) {
	originX := 0
	originY := int(m.GetHeight()) * -s.Client.AnimationsConfig.YStep.Y
	originX += y * s.Client.AnimationsConfig.YStep.X
	originY += y * s.Client.AnimationsConfig.YStep.Y
	originX += x * ctx.tileWidth
	originY += z * ctx.tileHeight

	indexZ := z
	indexX := x
	indexY := y

	targetZ = (indexZ * int(m.GetHeight()) * int(m.GetWidth())) + (int(m.GetDepth()) * indexY) - (indexX)

	// Calculate our scaled pixel position at which to render.
	targetX = int(float64(originX)*ctx.scale) + 100
	targetY = int(float64(originY)*ctx.scale) + 100
	return
}

// RenderObject renders a given Object within a DynamicMap.
func (s *Game) RenderObject(ctx RenderContext, viewObject *world.Object, o *world.Object, m *world.DynamicMap, dt time.Duration, uiMessages *BatchMessages) {
	if o != viewObject {
		if o.Element != nil {
			if o.Missing && o.WasMissing {
				return
			}
			if o.Missing && !o.WasMissing {
				o.WasMissing = true
				uiMessages.add(ui.BatchUpdateMessage{
					Target: o.Element,
					Update: ui.UpdateHidden(true),
				})
				return
			}
			if !o.Missing && o.WasMissing {
				o.WasMissing = false
				uiMessages.add(ui.BatchUpdateMessage{
					Target: o.Element,
					Update: ui.UpdateHidden(false),
				})
			}
		}
	}
	// Bail if there are no frames yet. FIXME: This should still show _something_
	if len(o.Face.Frames) == 0 {
		return
	}
	o.Process(dt)

	// Get and cache our render position.
	if o.Changed {
		o.RenderX, o.RenderY, o.RenderZ = s.GetRenderPosition(ctx, m, o.Y, o.X, o.Z)
		o.RenderZ += o.Index
		o.RecalculateFinalRender = true
	}

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

		offsetX += o.AdjustX
		offsetY += o.AdjustY

		// Set animation frame offsets.
		offsetX += int(o.Frame.X)
		offsetY += int(o.Frame.Y)

		// Adjust our target position.
		x := o.RenderX + int(float64(offsetX)*ctx.scale)
		y := o.RenderY + int(float64(offsetY)*ctx.scale)

		// Calculate our scaled pixel position at which to render.
		w = ctx.tileWidthScaled
		h = ctx.tileHeightScaled

		o.FinalRenderOffsetX = offsetX
		o.FinalRenderOffsetY = offsetY
		o.FinalRenderX = x
		o.FinalRenderY = y
		o.FinalRenderW = w
		o.FinalRenderH = h
	}

	// Get/create our shadow position, if we should.
	if o.HasShadow {
		s.RenderObjectShadows(ctx, o, m, o.FinalRenderOffsetX, o.FinalRenderOffsetY, o.FinalRenderW, o.FinalRenderH, uiMessages)
	}

	s.RenderObjectImage(ctx, o, m, o.Frame, o.FinalRenderX, o.FinalRenderY, o.RenderZ, o.FinalRenderW, o.FinalRenderH, uiMessages)
	return
}

func (s *Game) RenderObjectImage(ctx RenderContext, o *world.Object, m *world.DynamicMap, frame *data.AnimationFrame, x, y, zIndex, w, h int, uiMessages *BatchMessages) {
	if o.Image == nil {
		var err error
		o.Image, err = s.Client.DataManager.GetCachedImage(frame.ImageID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}

	if o.Element == nil {
		if o.Image != nil {
			bounds := o.Image.Bounds()
			w = int(float64(bounds.Max.X) * ctx.scale)
			h = int(float64(bounds.Max.Y) * ctx.scale)
			if (o.H > 1 || o.D > 1) && bounds.Max.Y > ctx.tileHeight {
				y -= h - ctx.tileHeightScaled
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
						// Ignore elements with an alpha less than or equal 0.1.
						if o.Element.GetStyle().Alpha.Value <= 0.1 {
							return true
						}
						// Ignore elements that are blocks or tiles.
						// TODO: Ignore if shift is held.
						if o.Type == cdata.ArchetypeBlock.AsUint8() || o.Type == cdata.ArchetypeTile.AsUint8() {
							return true
						}
						if o.Element.PixelHit(x, y) {
							s.inputChan <- elements.FocusObjectEvent{ID: o.ID}
							return false
						}
						return true
					},
					OnMouseMove: func(x int32, y int32) bool {
						if o.Element.GetStyle().Alpha.Value <= 0.1 {
							return true
						}
						// Ignore elements that are blocks or tiles.
						if o.Type == cdata.ArchetypeBlock.AsUint8() || o.Type == cdata.ArchetypeTile.AsUint8() {
							return true
						}
						if o.Element.PixelHit(x, y) {
							s.inputChan <- elements.HoverObjectEvent{ID: o.ID}
						} else {
							s.inputChan <- elements.UnhoverObjectEvent{ID: o.ID}
						}
						return true
					},
					OnMouseOut: func(x int32, y int32) bool {
						// Always unhover if the mouse leaves the object.
						s.inputChan <- elements.UnhoverObjectEvent{ID: o.ID}
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
							s.inputChan <- elements.FocusObjectEvent{ID: o.ID}
							return false
						}
						return true
					},
				},
			})
		}
		uiMessages.add(ui.BatchAdoptMessage{
			Parent: &s.MapWindow.Container,
			Target: o.Element,
		})
	} else {
		if o.Image != nil {
			if o.UnblockedChange {
				if o.Unblocked {
					uiMessages.add(ui.BatchUpdateMessage{
						Target: o.Element,
						Update: ui.UpdateAlpha(0.1),
					})
				} else {
					uiMessages.add(ui.BatchUpdateMessage{
						Target: o.Element,
						Update: ui.UpdateAlpha(1.0),
					})
				}
				o.UnblockedChange = false
			}
			if o.VisibilityChange {
				if o.Visible {
					uiMessages.add(ui.BatchUpdateMessage{
						Target: o.Element,
						Update: ui.UpdateGrayscale(false),
					})
				} else {
					uiMessages.add(ui.BatchUpdateMessage{
						Target: o.Element,
						Update: ui.UpdateGrayscale(true),
					})
				}
				o.VisibilityChange = false
			}
			if o.LightingChange {
				uiMessages.add(ui.BatchUpdateMessage{
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
				bounds := o.Image.Bounds()
				w = int(float64(bounds.Max.X) * ctx.scale)
				h = int(float64(bounds.Max.Y) * ctx.scale)

				var sw, sh float64
				sw = float64(w)
				sh = float64(h)
				if o.Squeezing {
					sw = math.Max(float64(w-w/4), float64(ctx.tileWidthScaled))
				}
				if o.Crouching {
					sh = math.Max(float64(h-h/3), float64(ctx.tileHeightScaled))
				}

				if (o.H > 1 || o.D > 1) && bounds.Max.Y > ctx.tileHeight {
					y -= int(sh) - ctx.tileHeightScaled
				}

				uiMessages.add(ui.BatchUpdateMessage{
					Target: o.Element,
					Update: ui.UpdateDimensions{
						X: ui.Number{Value: float64(x)},
						Y: ui.Number{Value: float64(y)},
						W: ui.Number{Value: sw},
						H: ui.Number{Value: sh},
					},
				})
				uiMessages.add(ui.BatchUpdateMessage{
					Target: o.Element,
					Update: ui.UpdateZIndex{Number: ui.Number{Value: float64(zIndex)}},
				})
				o.Changed = false
			}
			// Only update the image if the image ID has changed.
			if o.FrameImageID != frame.ImageID {
				o.FrameImageID = frame.ImageID
				uiMessages.add(ui.BatchUpdateMessage{
					Target: o.Element,
					Update: ui.UpdateImageID(o.FrameImageID),
				})
			}
		}
	}
	return
}

func (s *Game) RenderObjectShadows(ctx RenderContext, o *world.Object, m *world.DynamicMap, offsetX, offsetY, w, h int, uiMessages *BatchMessages) {
	// TODO: We should probably slice up an object's shadows based upon its width and depth. This will probably require using polygons unless SDL_gfx can clip rendered ellipses. Or, perhaps, use SDL_gfx's pie drawing calls for each shadow quadrant?
	sy, sx, sz := s.world.GetObjectShadowPosition(o)

	x, y, zIndex := s.GetRenderPosition(ctx, m, sy, sx, sz)
	// TODO: Fix shadows so they have a higher zIndex than z+1, but only for y of the same.
	zIndex--

	// Adjust our target position.
	x += int(float64(offsetX) * ctx.scale)
	y += int((float64(offsetY) + float64(o.D)) * ctx.scale)

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
		uiMessages.add(ui.BatchAdoptMessage{
			Parent: &s.MapWindow.Container,
			Target: o.ShadowElement,
		})
	} else {
		if o.Changed {
			uiMessages.add(ui.BatchUpdateMessage{
				Target: o.ShadowElement,
				Update: ui.UpdateDimensions{
					X: ui.Number{Value: float64(x)},
					Y: ui.Number{Value: float64(y)},
					W: ui.Number{Value: float64(w)},
					H: ui.Number{Value: float64(h)},
				},
			})

			uiMessages.add(ui.BatchUpdateMessage{
				Target: o.ShadowElement,
				Update: ui.UpdateZIndex{Number: ui.Number{Value: float64(zIndex)}},
			})

		}
	}
	return
}
