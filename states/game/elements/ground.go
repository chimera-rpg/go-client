package elements

import (
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

type GroundModeWindow struct {
	mode         GroundMode
	objects      [][][][]*world.Object
	Container    ui.Container
	toggleButton ui.ElementI
}

func (g *GroundModeWindow) Setup(style string, inputChan chan interface{}) (ui.Container, error) {
	g.Container.Setup(ui.ContainerConfig{
		Value: "Ground",
		Style: style,
	})
	g.toggleButton = ui.NewButtonElement(ui.ButtonElementConfig{
		Value: g.mode.String(),
		Style: `
			X 0
			Y 0
			W 64
			MinH 20
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
	g.Container.GetAdoptChannel() <- g.toggleButton

	return g.Container, nil
}

func (g *GroundModeWindow) SyncMode(mode GroundMode) {
	g.mode = mode
	g.toggleButton.GetUpdateChannel() <- ui.UpdateValue{Value: mode.String()}
}

func (g *GroundModeWindow) SyncObjects() {
	// TODO: iterate over g.objects, add/remove Y, X, Z, and tiles container quandrants to fit, then add or replace image elements, etc. as needed.
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
	// 1. Collect a 3D slice of all notable objects in range.
	var objects [][][][]*world.Object
	for ys := minY; ys < maxY; ys++ {
		y := ys + reach
		objects = append(objects, make([][][]*world.Object, 0))
		for xs := minX; xs < maxX; xs++ {
			x := xs + reach
			objects[y] = append(objects[y], make([][]*world.Object, 0))
			for zs := minZ; zs < maxZ; zs++ {
				z := zs - minZ
				objects[y][x] = append(objects[y][x], make([]*world.Object, 0))
				if t := m.GetTile(vo.Y+ys, vo.X+xs, vo.Z+zs); t != nil {
					for _, o := range t.Objects() {
						if slices.Contains(typeFilter, o.Type) {
							objects[y][x][z] = append(objects[y][x][z], o)
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
