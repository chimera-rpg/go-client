package elements

import (
	"fmt"
	"image/color"

	"github.com/chimera-rpg/go-client/config"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-client/world"
	"github.com/chimera-rpg/go-server/data"
	"golang.org/x/exp/slices"
)

// GroundMode is the type for indicating the current ground mode.
type GroundMode int

const (
	GroundModeNearby = iota
	GroundModeUnderfoot
)

func (g GroundMode) String() string {
	if g == GroundModeUnderfoot {
		return "underfoot"
	}
	return "nearby"
}

// GroundModeChangeEvent is used to change the ground/nearby items mode.
type GroundModeChangeEvent struct {
	ID   string
	Mode GroundMode
}

// GroundModeComboEvent toggles the aggregate item collection.
type GroundModeComboEvent struct {
	ID string
}

type ObjectReference struct {
	count     int
	object    *world.Object
	objectIDs []uint32
}

type ObjectContainer struct {
	container *ui.Container
	image     ui.ElementI
	count     ui.ElementI
	lastCount int
	hidden    bool
}

// ContainerWindow represents any sort of listing of objects that are distinct from the map representation. The same structure is used to represent the nearby items list, the player's inventories, and opened containers such as chests.
type ContainerWindow struct {
	Type             ContainerType
	ID               string
	Config           *config.ContainerConfig
	game             game
	objects          []ObjectReference
	Container        *ui.Container
	objectsList      *ui.Container
	objectContainers []ObjectContainer
	nearbyButton     ui.ElementI
	underfootButton  ui.ElementI
	aggregateButton  ui.ElementI
	focusedContainer *ui.Container
}

// ContainerType is used to control what type of container this container represents.
type ContainerType int

const (
	// ContainerInventoryType represents the player's base inventory.
	ContainerInventoryType ContainerType = iota
	// ContainerGroundType represents what the player can see nearby.
	ContainerGroundType
)

// ContainerWindowConfig is a configuration structure used to determine how a container should look and function.
type ContainerWindowConfig struct {
	ID    string // ID is used as the key for storing the container's configuration to disk.
	Type  ContainerType
	Style string
}

// Setup sets up the container with the given game, config, and input channel.
func (g *ContainerWindow) Setup(game game, cfg ContainerWindowConfig, inputChan chan interface{}) (*ui.Container, error) {
	g.game = game
	g.Type = cfg.Type
	g.ID = cfg.ID
	g.Config = game.Config().Game.GetContainerConfig(g.ID)
	var err error
	g.Container, err = ui.NewContainerElement(ui.ContainerConfig{
		Value: "Container",
		Style: cfg.Style,
	})
	if err != nil {
		return nil, err
	}
	g.objectsList, err = ui.NewContainerElement(ui.ContainerConfig{
		Style: `
			Y 10%
			W 100%
			H 90%
		`,
	})
	if err != nil {
		return nil, err
	}
	if cfg.Type == ContainerGroundType {
		g.nearbyButton = ui.NewButtonElement(ui.ButtonElementConfig{
			Value: "nearby",
			Style: `
			X 0
			Y 0
			W 64
			H 10%
		`,
			NoFocus: true,
			Events: ui.Events{
				OnMouseButtonUp: func(button uint8, x, y int32) bool {
					inputChan <- GroundModeChangeEvent{
						ID:   g.ID,
						Mode: GroundModeNearby,
					}
					return false
				},
			},
		})
		g.underfootButton = ui.NewButtonElement(ui.ButtonElementConfig{
			Value: `underfoot`,
			Style: `
			X 64
			Y 0
			W 96
			H 10%
		`,
			NoFocus: true,
			Events: ui.Events{
				OnMouseButtonUp: func(button uint8, x, y int32) bool {
					inputChan <- GroundModeChangeEvent{
						ID:   g.ID,
						Mode: GroundModeUnderfoot,
					}
					return false
				},
			},
		})
		g.Container.GetAdoptChannel() <- g.nearbyButton
		g.Container.GetAdoptChannel() <- g.underfootButton
	}
	g.aggregateButton = ui.NewButtonElement(ui.ButtonElementConfig{
		Value: "C",
		Style: `
			X 0
			Y 0
			Origin Right
			W 32
			H 10%
		`,
		NoFocus: true,
		Events: ui.Events{
			OnMouseButtonUp: func(button uint8, x, y int32) bool {
				inputChan <- GroundModeComboEvent{
					ID: g.ID,
				}
				return false
			},
		},
	})

	g.Container.GetAdoptChannel() <- g.objectsList.This
	g.Container.GetAdoptChannel() <- g.aggregateButton

	g.SyncMode(GroundMode(g.Config.Mode))
	g.RefreshCombo()

	game.HookEvent(GroundModeComboEvent{}, func(e interface{}) {
		if e.(GroundModeComboEvent).ID != g.ID {
			return
		}
		g.ToggleCombo()
		g.Refresh()
	})
	game.HookEvent(GroundModeChangeEvent{}, func(e interface{}) {
		if e.(GroundModeChangeEvent).ID != g.ID {
			return
		}
		g.SyncMode(e.(GroundModeChangeEvent).Mode)
		g.Refresh()
	})
	game.HookEvent(FocusObjectEvent{}, func(e interface{}) {
		g.RefreshFocus()
	})

	return g.Container, nil
}

func (g *ContainerWindow) RefreshFocus() {
	found := false
	for ori, or := range g.objects {
		for _, id := range or.objectIDs {
			if id == g.game.FocusedObjectID() {
				if g.focusedContainer != nil {
					g.focusedContainer.GetUpdateChannel() <- ui.UpdateBackgroundColor{
						R: 0,
						G: 0,
						B: 255,
						A: 32,
					}
				}
				g.focusedContainer = g.objectContainers[ori].container

				g.focusedContainer.GetUpdateChannel() <- ui.UpdateBackgroundColor{
					R: 0,
					G: 255,
					B: 0,
					A: 64,
				}
				found = true
			}
		}
	}
	if !found {
		if g.focusedContainer != nil {
			g.focusedContainer.GetUpdateChannel() <- ui.UpdateBackgroundColor{
				R: 0,
				G: 0,
				B: 255,
				A: 32,
			}
			g.focusedContainer = nil
		}
	}
}

func (g *ContainerWindow) SyncMode(mode GroundMode) {
	// FIXME: Load these from some sort of passed in Stylesheet global
	inactiveColor := color.NRGBA{64, 64, 111, 128}
	activeColor := color.NRGBA{139, 139, 186, 128}
	g.Config.Mode = int(mode)
	if g.Type == ContainerGroundType {
		if mode == GroundModeNearby {
			g.nearbyButton.GetUpdateChannel() <- ui.UpdateBackgroundColor(activeColor)
			g.underfootButton.GetUpdateChannel() <- ui.UpdateBackgroundColor(inactiveColor)
		} else if mode == GroundModeUnderfoot {
			g.underfootButton.GetUpdateChannel() <- ui.UpdateBackgroundColor(activeColor)
			g.nearbyButton.GetUpdateChannel() <- ui.UpdateBackgroundColor(inactiveColor)
		}
	}
}

func (g *ContainerWindow) ToggleCombo() {
	g.Config.Aggregate = !g.Config.Aggregate
	g.RefreshCombo()
}

func (g *ContainerWindow) RefreshCombo() {
	// FIXME: Load these from some sort of passed in Stylesheet global
	inactiveColor := color.NRGBA{64, 64, 111, 128}
	activeColor := color.NRGBA{139, 139, 186, 128}

	if g.Config.Aggregate {
		g.aggregateButton.GetUpdateChannel() <- ui.UpdateBackgroundColor(activeColor)
	} else {
		g.aggregateButton.GetUpdateChannel() <- ui.UpdateBackgroundColor(inactiveColor)
	}
}

func (g *ContainerWindow) SyncObjects() {
	batchMessages := make([]ui.BatchMessage, 0)
	w := 48
	h := 48
	if len(g.objectContainers) > len(g.objects) {
		for i := len(g.objects); i < len(g.objectContainers); i++ {
			if !g.objectContainers[i].hidden {
				batchMessages = append(batchMessages, ui.BatchUpdateMessage{
					Target: g.objectContainers[i].container,
					Update: ui.UpdateHidden(true),
				})
				g.objectContainers[i].hidden = true
			}
		}
	}
	if len(g.objectContainers) < len(g.objects) {
		for i := len(g.objectContainers); i < len(g.objects); i++ {
			func(i int) {
				el, _ := ui.NewContainerElement(ui.ContainerConfig{
					Style: fmt.Sprintf(`
					W %d
					H %d
					BackgroundColor 0 0 255 32
				`, w, h),
					Events: ui.Events{
						OnMouseButtonUp: func(button uint8, x, y int32) bool {
							g.game.InputChan() <- FocusObjectEvent{
								ID: g.objects[i].objectIDs[0],
							}
							return false
						},
					},
				})
				box := ui.NewPrimitiveElement(ui.PrimitiveElementConfig{
					Shape: ui.RectangleShape,
					Style: `
					X 0
					Y 0
					W 100%
					H 100%
					OutlineColor 0 0 0 255
				`,
				})
				img := ui.NewImageElement(ui.ImageElementConfig{
					Style: `
					Origin CenterX CenterY
					X 50%
					Y 50%
				`,
				})
				count := ui.NewTextElement(ui.TextElementConfig{
					Value: "egg",
					Style: `
					X 0
					Y 0
					PaddingTop -4
					PaddingLeft 1
					Resize ToContent
					OutlineColor 255 255 255 128
				`,
				})
				batchMessages = append(batchMessages, ui.BatchAdoptMessage{
					Parent: g.objectsList,
					Target: el,
				})

				g.objectContainers = append(g.objectContainers, ObjectContainer{
					container: el,
					image:     img,
					count:     count,
				})
				batchMessages = append(batchMessages, ui.BatchAdoptMessage{
					Parent: el,
					Target: box,
				})
				batchMessages = append(batchMessages, ui.BatchAdoptMessage{
					Parent: el,
					Target: img,
				})
				batchMessages = append(batchMessages, ui.BatchAdoptMessage{
					Parent: el,
					Target: count,
				})
			}(i)
		}
	}
	x := 0
	y := 0
	// row and col are used to overlap the items by 1 pixel so their borders overlap.
	row := 0
	col := 0
	maxWidth := int(g.objectsList.GetWidth())
	for i := range g.objects {
		c := &(g.objectContainers[i])
		if x+w >= maxWidth {
			x = 0
			y += h
			row++
			col = 0
		}

		if c.hidden {
			batchMessages = append(batchMessages, ui.BatchUpdateMessage{
				Target: c.container,
				Update: ui.UpdateHidden(false),
			})
			c.hidden = false
		}

		batchMessages = append(batchMessages, ui.BatchUpdateMessage{
			Target: c.container,
			Update: ui.UpdateX{Number: ui.Number{Value: float64(x - col)}},
		})
		batchMessages = append(batchMessages, ui.BatchUpdateMessage{
			Target: c.container,
			Update: ui.UpdateY{Number: ui.Number{Value: float64(y - row)}},
		})

		if g.objects[i].object.FrameImageID > 0 {
			bounds := g.objects[i].object.Image.Bounds()
			batchMessages = append(batchMessages, ui.BatchUpdateMessage{
				Target: c.image,
				Update: ui.UpdateDimensions{
					X: c.image.GetStyle().X,
					Y: c.image.GetStyle().Y,
					W: ui.Number{Value: float64(bounds.Dx() * 2)},
					H: ui.Number{Value: float64(bounds.Dy() * 2)},
				},
			})

			batchMessages = append(batchMessages, ui.BatchUpdateMessage{
				Target: c.image,
				Update: ui.UpdateImageID(g.objects[i].object.FrameImageID),
			})
		}

		// lastCount is an optimization to prevent unnecessarily changing the text value which is expensive due to no font atlas being used.
		if c.lastCount != g.objects[i].count {
			if g.objects[i].count == 1 {
				batchMessages = append(batchMessages, ui.BatchUpdateMessage{
					Target: c.count,
					Update: ui.UpdateValue{Value: ""},
				})
			} else {
				batchMessages = append(batchMessages, ui.BatchUpdateMessage{
					Target: c.count,
					Update: ui.UpdateValue{Value: fmt.Sprintf("%d", g.objects[i].count)},
				})
			}
			c.lastCount = g.objects[i].count
		}

		x += w
		col++
	}

	if len(batchMessages) > 0 {
		ui.GlobalInstance.RootWindow.BatchChannel <- batchMessages
	}
	//g.Container.GetUpdateChannel() <- ui.UpdateDirt(true)
}

// Refresh assigns the view to a slice of tiles.
func (g *ContainerWindow) Refresh() {
	w := g.game.World()
	vo := w.GetViewObject()
	m := w.GetCurrentMap()

	if g.Type == ContainerInventoryType {
		var objects []ObjectReference
		if c := g.game.World().GetContainer(0); c != nil {
			for _, oID := range c.ObjectIDs {
				if o := g.game.World().GetObject(oID); o != nil {
					if g.Config.Aggregate {
						found := false
						for oi, or := range objects {
							// FIXME: It doesn't seem correct to use animation and face IDs to identify same object types. Perhaps the archetype's underlying ID should also be passed with the standard object creation network data?
							if or.object.AnimationID == o.AnimationID && or.object.FaceID == o.FaceID {
								objects[oi].count++
								objects[oi].objectIDs = append(objects[oi].objectIDs, o.ID)
								found = true
								break
							}
						}
						if !found {
							objects = append(objects, ObjectReference{
								object:    o,
								count:     1,
								objectIDs: []uint32{o.ID},
							})
						}
					} else {
						objects = append(objects, ObjectReference{
							object:    o,
							count:     1,
							objectIDs: []uint32{o.ID},
						})
					}
				}
			}
		}
		g.objects = objects
	} else if g.Type == ContainerGroundType {
		// Default type filter.
		typeFilter := []uint8{data.ArchetypeEquipable.AsUint8(), data.ArchetypeItem.AsUint8(), data.ArchetypeGeneric.AsUint8(), data.ArchetypeFood.AsUint8()}

		// Use our reach cube per default.
		cube := w.ReachCube
		reachX := int(vo.Reach)
		reachY := int(vo.Reach)
		reachZ := int(vo.Reach)

		// Otherwise use the bottom-1 of our intersect cube.
		if g.Config.Mode == GroundModeUnderfoot {
			// Reassign type filter if we're looking underfoot.
			typeFilter = append(typeFilter, data.ArchetypeBlock.AsUint8(), data.ArchetypeTile.AsUint8())
			cube = w.IntersectCube[:1]
			cube = append(cube, cube...)
			reachX = 0
			reachZ = 0
			reachY = 1
		}

		// 1. Collect a slice of all notable objects in range.
		var objects []ObjectReference
		for ys := range cube {
			for xs := range cube[ys] {
				for zs := range cube[ys][xs] {
					if t := m.GetTile(vo.Y+ys-reachY, vo.X+xs-reachX, vo.Z+zs-reachZ); t != nil {
						for _, o := range t.Objects() {
							if slices.Contains(typeFilter, o.Type) {
								if g.Config.Aggregate {
									found := false
									for oi, or := range objects {
										// FIXME: It doesn't seem correct to use animation and face IDs to identify same object types. Perhaps the archetype's underlying ID should also be passed with the standard object creation network data?
										if or.object.AnimationID == o.AnimationID && or.object.FaceID == o.FaceID {
											objects[oi].count++
											objects[oi].objectIDs = append(objects[oi].objectIDs, o.ID)
											found = true
											break
										}
									}
									if !found {
										objects = append(objects, ObjectReference{
											object:    o,
											count:     1,
											objectIDs: []uint32{o.ID},
										})
									}
								} else {
									objects = append(objects, ObjectReference{
										object:    o,
										count:     1,
										objectIDs: []uint32{o.ID},
									})
								}
							}
						}

					}
				}
			}
		}
		g.objects = objects
	}
	// 2. Synchronize objects.
	g.SyncObjects()
	// 3. Refresh focus in case something went out of range.
	g.RefreshFocus()
}
