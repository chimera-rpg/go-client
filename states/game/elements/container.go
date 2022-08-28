package elements

import (
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-client/world"
)

type ContainerWindow struct {
	game             game
	container        *ui.Container
	aggregateButton  ui.ElementI
	focusedContainer *ui.Container
	objectsList      *ui.Container
	objects          []ObjectReference
	objectContainers []ObjectContainer
}

func (c *ContainerWindow) Setup(game game, style string, inputChan chan interface{}) (*ui.Container, error) {
	c.game = game
	var err error
	c.container, err = ui.NewContainerElement(ui.ContainerConfig{
		Value: "Container",
		Style: style,
	})
	if err != nil {
		return nil, err
	}
	c.objectsList, err = ui.NewContainerElement(ui.ContainerConfig{
		Style: `
			Y 20
			W 100%
			H 100%
		`,
	})
	if err != nil {
		return nil, err
	}
	c.aggregateButton = ui.NewButtonElement(ui.ButtonElementConfig{
		Value: "C",
		Style: `
			X 0
			Y 0
			Origin Right
			W 32
			H 20
		`,
		NoFocus: true,
		Events: ui.Events{
			OnMouseButtonUp: func(button uint8, x, y int32) bool {
				//inputChan <- GroundModeComboEvent{}
				return false
			},
		},
	})
	c.container.GetAdoptChannel() <- c.objectsList.This
	c.container.GetAdoptChannel() <- c.aggregateButton

	return c.container, nil
}

// SyncTo syncs the container's object references to the given objects slice.
func (c *ContainerWindow) SyncTo(objects []*world.Object) {
	//
}

// Refresh refreshes the window to show focused, change images, etc.
func (c *ContainerWindow) Refresh() {
}
