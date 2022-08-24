package elements

import (
	"fmt"

	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-client/world"
	"github.com/chimera-rpg/go-common/data"
	"golang.org/x/exp/slices"
)

// GroundMode is the type for indicating the current ground mode.
type GroundMode int

const (
	GroundModeNearby = iota
	GroundModeExact
)

func (g GroundMode) String() string {
	if g == GroundModeExact {
		return "exact"
	}
	return "nearby"
}

// GroundModeEvent is used to change the ground/nearby items mode.
type GroundModeEvent struct {
	Mode GroundMode
}

type ObjectContainer struct {
	container *ui.Container
	image     ui.ElementI
}

type GroundModeWindow struct {
	mode             GroundMode
	objects          []*world.Object
	Container        ui.Container
	objectsList      ui.Container
	objectContainers []ObjectContainer
	toggleButton     ui.ElementI
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
	g.toggleButton = ui.NewButtonElement(ui.ButtonElementConfig{
		Value: g.mode.String(),
		Style: `
			X 0
			Y 0
			W 64
			H 10%
		`,
		NoFocus: true,
		Events: ui.Events{
			OnMouseButtonUp: func(button uint8, x, y int32) bool {
				var mode GroundMode
				if g.mode == GroundModeNearby {
					mode = GroundModeExact
				} else {
					mode = GroundModeNearby
				}
				inputChan <- GroundModeEvent{
					Mode: mode,
				}
				return false
			},
		},
	})
	g.Container.GetAdoptChannel() <- g.objectsList.This
	g.Container.GetAdoptChannel() <- g.toggleButton

	return g.Container, nil
}

func (g *GroundModeWindow) SyncMode(mode GroundMode) {
	g.mode = mode
	g.toggleButton.GetUpdateChannel() <- ui.UpdateValue{Value: mode.String()}
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
					Resize ToContent
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
			g.objectContainers = append(g.objectContainers, ObjectContainer{
				container: el,
				image:     img,
			})
			el.GetAdoptChannel() <- box
			el.GetAdoptChannel() <- img
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

		if g.objects[i].FrameImageID > 0 {
			bounds := g.objects[i].Image.Bounds()
			c.image.GetUpdateChannel() <- ui.UpdateDimensions{
				X: c.image.GetStyle().X,
				Y: c.image.GetStyle().Y,
				W: ui.Number{Value: float64(bounds.Dx() * 2)},
				H: ui.Number{Value: float64(bounds.Dy() * 2)},
			}
			c.image.GetUpdateChannel() <- ui.UpdateImageID(g.objects[i].FrameImageID)
		}

		// TODO: If focusedObjectID == g.objects[i].ID, set background

		x += w
		col++
	}

	g.Container.GetUpdateChannel() <- ui.UpdateDirt(true)
}

// Refresh assigns the view to a 3D slice of tiles.
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
	//
	typeFilter := []uint8{data.ArchetypeArmor.AsUint8(), data.ArchetypeWeapon.AsUint8(), data.ArchetypeItem.AsUint8(), data.ArchetypeGeneric.AsUint8(), data.ArchetypeShield.AsUint8(), data.ArchetypeFood.AsUint8()}
	// 1. Collect a slice of all notable objects in range.
	var objects []*world.Object
	for xs := minX; xs < maxX; xs++ {
		for zs := minZ; zs < maxZ; zs++ {
			for ys := minY; ys < maxY; ys++ {
				if t := m.GetTile(vo.Y+ys, vo.X+xs, vo.Z+zs); t != nil {
					for _, o := range t.Objects() {
						if slices.Contains(typeFilter, o.Type) {
							objects = append(objects, o)
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
