package elements

import (
	"fmt"
	"image/color"

	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-client/world"
	"github.com/chimera-rpg/go-common/data"
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
	Mode GroundMode
}

// GroundModeComboEvent toggles the aggregate item collection.
type GroundModeComboEvent struct {
}

type ObjectReference struct {
	count  int
	object *world.Object
}

type ObjectContainer struct {
	container *ui.Container
	image     ui.ElementI
	count     ui.ElementI
}

type GroundModeWindow struct {
	mode             GroundMode
	aggregate        bool
	objects          []ObjectReference
	Container        ui.Container
	objectsList      ui.Container
	objectContainers []ObjectContainer
	nearbyButton     ui.ElementI
	underfootButton  ui.ElementI
	aggregateButton  ui.ElementI
}

func (g *GroundModeWindow) Setup(style string, inputChan chan interface{}) (ui.Container, error) {
	g.Container.Setup(ui.ContainerConfig{
		Value: "Ground",
		Style: style,
	})
	g.objectsList.Setup(ui.ContainerConfig{
		Style: `
			Y 10%
			W 100%
			H 90%
		`,
	})
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
					Mode: GroundModeUnderfoot,
				}
				return false
			},
		},
	})
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
				inputChan <- GroundModeComboEvent{}
				return false
			},
		},
	})

	g.Container.GetAdoptChannel() <- g.objectsList.This
	g.Container.GetAdoptChannel() <- g.nearbyButton
	g.Container.GetAdoptChannel() <- g.underfootButton
	g.Container.GetAdoptChannel() <- g.aggregateButton

	g.SyncMode(g.mode)
	g.ToggleCombo()

	return g.Container, nil
}

func (g *GroundModeWindow) SyncMode(mode GroundMode) {
	// FIXME: Load these from some sort of passed in Stylesheet global
	inactiveColor := color.NRGBA{64, 64, 111, 128}
	activeColor := color.NRGBA{139, 139, 186, 128}
	g.mode = mode
	if g.mode == GroundModeNearby {
		g.nearbyButton.GetUpdateChannel() <- ui.UpdateBackgroundColor(activeColor)
		g.underfootButton.GetUpdateChannel() <- ui.UpdateBackgroundColor(inactiveColor)
	} else if g.mode == GroundModeUnderfoot {
		g.underfootButton.GetUpdateChannel() <- ui.UpdateBackgroundColor(activeColor)
		g.nearbyButton.GetUpdateChannel() <- ui.UpdateBackgroundColor(inactiveColor)
	}
}

func (g *GroundModeWindow) ToggleCombo() {
	// FIXME: Load these from some sort of passed in Stylesheet global
	inactiveColor := color.NRGBA{64, 64, 111, 128}
	activeColor := color.NRGBA{139, 139, 186, 128}

	g.aggregate = !g.aggregate
	if g.aggregate {
		g.aggregateButton.GetUpdateChannel() <- ui.UpdateBackgroundColor(activeColor)
	} else {
		g.aggregateButton.GetUpdateChannel() <- ui.UpdateBackgroundColor(inactiveColor)
	}
}

func (g *GroundModeWindow) SyncObjects() {
	w := 48
	h := 48
	if len(g.objectContainers) > len(g.objects) {
		for i := len(g.objects); i < len(g.objectContainers); i++ {
			g.objectContainers[i].container.GetDestroyChannel() <- true
		}
		g.objectContainers = g.objectContainers[:len(g.objects)]
	}
	if len(g.objectContainers) < len(g.objects) {
		for i := len(g.objectContainers); i < len(g.objects); i++ {
			el, _ := ui.NewContainerElement(ui.ContainerConfig{
				Style: fmt.Sprintf(`
					W %d
					H %d
					BackgroundColor 0 0 255 32
				`, w, h),
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
			g.objectContainers = append(g.objectContainers, ObjectContainer{
				container: el,
				image:     img,
				count:     count,
			})
			el.GetAdoptChannel() <- box
			el.GetAdoptChannel() <- img
			el.GetAdoptChannel() <- count
			g.objectsList.GetAdoptChannel() <- el.This
		}
	}
	x := 0
	y := 0
	row := 0
	col := 0
	maxWidth := int(g.objectsList.GetWidth())
	for i, c := range g.objectContainers {
		if x+w >= maxWidth {
			x = 0
			y += h
			row++
			col = 0
		}

		c.container.GetUpdateChannel() <- ui.UpdateX{Number: ui.Number{Value: float64(x - col)}}
		c.container.GetUpdateChannel() <- ui.UpdateY{Number: ui.Number{Value: float64(y - row)}}

		if g.objects[i].object.FrameImageID > 0 {
			bounds := g.objects[i].object.Image.Bounds()
			c.image.GetUpdateChannel() <- ui.UpdateDimensions{
				X: c.image.GetStyle().X,
				Y: c.image.GetStyle().Y,
				W: ui.Number{Value: float64(bounds.Dx() * 2)},
				H: ui.Number{Value: float64(bounds.Dy() * 2)},
			}
			c.image.GetUpdateChannel() <- ui.UpdateImageID(g.objects[i].object.FrameImageID)
		}

		if g.objects[i].count == 1 {
			c.count.GetUpdateChannel() <- ui.UpdateValue{Value: ""}
		} else {
			c.count.GetUpdateChannel() <- ui.UpdateValue{Value: fmt.Sprintf("%d", g.objects[i].count)}
		}
		// TODO: If focusedObjectID == g.objects[i].ID, set background

		x += w
		col++
	}

	g.Container.GetUpdateChannel() <- ui.UpdateDirt(true)
}

// Refresh assigns the view to a slice of tiles.
func (g *GroundModeWindow) RefreshFromWorld(w *world.World) {
	vo := w.GetViewObject()
	m := w.GetCurrentMap()
	// FIXME: We need the view object's Reach value!
	reach := 2
	minY := -reach
	maxY := int(vo.H) + reach
	minX := -reach
	maxX := int(vo.W) + reach
	minZ := -reach
	if vo.D > 1 {
		minZ -= int(vo.D)
	}
	maxZ := reach

	// Default type filter.
	typeFilter := []uint8{data.ArchetypeArmor.AsUint8(), data.ArchetypeWeapon.AsUint8(), data.ArchetypeItem.AsUint8(), data.ArchetypeGeneric.AsUint8(), data.ArchetypeShield.AsUint8(), data.ArchetypeFood.AsUint8()}

	if g.mode == GroundModeUnderfoot {
		minY = -1
		maxY = 1
		minX = 0
		maxX = int(vo.W)
		minZ = 0
		maxZ = int(vo.D)
		// Reassign type filter if we're looking underfoot.
		typeFilter = append(typeFilter, data.ArchetypeBlock.AsUint8(), data.ArchetypeTile.AsUint8())
	}
	// 1. Collect a slice of all notable objects in range.
	var objects []ObjectReference
	for xs := minX; xs < maxX; xs++ {
		for zs := minZ; zs < maxZ; zs++ {
			for ys := minY; ys < maxY; ys++ {
				if t := m.GetTile(vo.Y+ys, vo.X+xs, vo.Z+zs); t != nil {
					for _, o := range t.Objects() {
						if slices.Contains(typeFilter, o.Type) {
							if g.aggregate {
								found := false
								for oi, or := range objects {
									// FIXME: It doesn't seem correct to use animation and face IDs to identify same object types. Perhaps the archetype's underlying ID should also be passed with the standard object creation network data?
									if or.object.AnimationID == o.AnimationID && or.object.FaceID == o.FaceID {
										objects[oi].count++
										found = true
										break
									}
								}
								if !found {
									objects = append(objects, ObjectReference{
										object: o,
										count:  1,
									})
								}
							} else {
								objects = append(objects, ObjectReference{
									object: o,
									count:  1,
								})
							}
						}
					}
				}
			}
		}
	}
	g.objects = objects
	// 2. Synchronize objects.
	g.SyncObjects()
}
