package elements

import (
	"fmt"

	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-common/network"
)

type InspectorWindow struct {
	game            game
	container       *ui.Container
	object          ObjectReference
	imageContainer  *ui.Container
	image           ui.ElementI
	count           ui.ElementI
	name            ui.ElementI
	description     ui.ElementI
	types           ui.ElementI
	extra           ui.ElementI
	focusedObjectID uint32
	inRange         bool
	blank           bool
}

func (w *InspectorWindow) Setup(game game, style string, inputChan chan interface{}) (*ui.Container, error) {
	w.game = game
	var err error
	w.container, err = ui.NewContainerElement(ui.ContainerConfig{
		Style: style,
	})
	if err != nil {
		return nil, err
	}

	w.imageContainer, err = ui.NewContainerElement(ui.ContainerConfig{
		Style: fmt.Sprintf(`
			W %d
			H %d
			BackgroundColor 32 32 32 255
		`, 64, 64),
	})
	if err != nil {
		return nil, err
	}

	w.image = ui.NewImageElement(ui.ImageElementConfig{
		Style: `
			Origin CenterX CenterY
			X 50%
			Y 50%
			OutlineColor 255 255 0 150
		`,
	})

	w.name = ui.NewTextElement(ui.TextElementConfig{
		Value: "",
		Style: `
			X 64
			Y 0
			PaddingLeft 2
			PaddingTop 2
			ForegroundColor 255 255 255 255
			OutlineColor 0 0 0 200
		`,
	})

	w.types = ui.NewTextElement(ui.TextElementConfig{
		Value: "",
		Style: `
			X 64
			Y 20
			PaddingLeft 2
			PaddingTop 2
			ForegroundColor 255 255 255 255
		`,
	})

	w.extra = ui.NewTextElement(ui.TextElementConfig{
		Value: "",
		Style: `
			X 64
			Y 40
			PaddingLeft 2
			PaddingTop 2
			ForegroundColor 255 255 255 255
		`,
	})

	w.container.GetAdoptChannel() <- w.name
	w.container.GetAdoptChannel() <- w.types
	w.container.GetAdoptChannel() <- w.extra
	w.container.GetAdoptChannel() <- w.imageContainer
	w.imageContainer.GetAdoptChannel() <- w.image

	w.game.HookEvent(FocusObjectEvent{}, func(e interface{}) {
		w.Refresh()
	})

	return w.container, err
}

func (w *InspectorWindow) Refresh() {
	o := w.game.World().GetObject(w.game.FocusedObjectID())
	if o == nil {
		if !w.blank {
			w.blank = true
			w.image.GetUpdateChannel() <- ui.UpdateHidden(true)
			w.name.GetUpdateChannel() <- ui.UpdateHidden(true)
			w.types.GetUpdateChannel() <- ui.UpdateHidden(true)
			w.extra.GetUpdateChannel() <- ui.UpdateHidden(true)
		}
	} else {
		if w.blank {
			w.image.GetUpdateChannel() <- ui.UpdateHidden(false)
			w.name.GetUpdateChannel() <- ui.UpdateHidden(false)
			w.types.GetUpdateChannel() <- ui.UpdateHidden(false)
			w.extra.GetUpdateChannel() <- ui.UpdateHidden(false)
		}
		w.blank = false
		var refresh bool
		vo := w.game.World().GetViewObject()
		if vo != nil {
			minY := vo.Y - int(vo.Reach)
			maxY := vo.Y + int(vo.H) + int(vo.Reach)
			minX := vo.X - int(vo.Reach)
			maxX := vo.X + int(vo.W) + int(vo.Reach)
			minZ := vo.Z - int(vo.Reach)
			maxZ := vo.Z + int(vo.D) + int(vo.Reach)

			if o.Y >= minY && o.Y < maxY && o.X >= minX && o.X < maxX && o.Z >= minZ && o.Z < maxZ {
				if !w.inRange {
					if len(o.Info) < 1 || !o.Info[0].Near {
						w.game.SendNetMessage(network.CommandInspect{
							ObjectID: w.game.FocusedObjectID(),
						})
					}
				}
				w.inRange = true
			} else {
				w.inRange = false
			}
		}
		if w.focusedObjectID != w.game.FocusedObjectID() {
			if !o.HasInfo {
				w.game.SendNetMessage(network.CommandInspect{
					ObjectID: w.game.FocusedObjectID(),
				})
			}
			refresh = true
		}
		if o.InfoChange || o.HasInfo {
			o.InfoChange = false
			refresh = true
		}
		if refresh {
			// Refresh image.
			if o.Image != nil {
				w.image.GetUpdateChannel() <- ui.UpdateImageID(o.FrameImageID)
				bounds := o.Image.Bounds()
				w.image.GetUpdateChannel() <- ui.UpdateDimensions{
					X: w.image.GetStyle().X,
					Y: w.image.GetStyle().Y,
					W: ui.Number{Value: float64(bounds.Dx() * 3)},
					H: ui.Number{Value: float64(bounds.Dy() * 3)},
				}
			}
			// Refresh information.
			name := "?"
			extra := ""
			typeString := ""
			useSlots := make(map[string]int)
			types := make([]string, 0)
			for _, info := range o.Info {
				if info.Name != "" {
					name = info.Name
				}
				// Use types.
				for _, t := range info.TypeHints {
					s := w.game.TypeHint(t)
					has := false
					for _, t2 := range types {
						if s == t2 {
							has = true
							break
						}
					}
					if !has {
						types = append(types, s)
					}
				}
				// Use slots
				for k, v := range info.Slots.Uses {
					useSlots[w.game.Slot(k)] = v
				}
			}
			if len(useSlots) > 0 {
				used := ""
				for s, i := range useSlots {
					used += fmt.Sprintf("%d %s", i, s)
				}
				extra += used
			}
			for i, v := range types {
				if i == 0 {
					typeString = v
				} else {
					typeString += ", " + v
				}
			}

			w.name.GetUpdateChannel() <- ui.UpdateValue{Value: name}
			w.types.GetUpdateChannel() <- ui.UpdateValue{Value: typeString}
			w.extra.GetUpdateChannel() <- ui.UpdateValue{Value: extra}
		}
	}
	w.focusedObjectID = w.game.FocusedObjectID()
}
