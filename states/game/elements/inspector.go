package elements

import "github.com/chimera-rpg/go-client/ui"

type InspectorWindow struct {
	container   *ui.Container
	object      ObjectReference
	image       ui.ElementI
	count       ui.ElementI
	name        ui.ElementI
	description ui.ElementI
}

func (w *InspectorWindow) Setup(style string, inputChan chan interface{}) (*ui.Container, error) {
	var err error
	w.container, err = ui.NewContainerElement(ui.ContainerConfig{
		Style: style,
	})
	if err != nil {
		return nil, err
	}

	return w.container, err
}
