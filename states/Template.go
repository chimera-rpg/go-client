package states

import (
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/ui"
)

// Template is a template state.
type Template struct {
	client.State
    TemplateWindow ui.Window
}

// Init our Template state.
func (s *Template) Init(v interface{}) (next client.StateI, nextArgs interface{}, err error) {
    var elButton ui.ElementI

	err = s.TemplateWindow.Setup(ui.WindowConfig{
		Value: "Selection",
		Style: `
			W 100%
			H 100%
			BackgroundColor 139 186 139 255
		`,
		Parent: s.Client.RootWindow,
	})

	elButton = ui.NewButtonElement(ui.ButtonElementConfig{
		Style: `
			Origin CenterX CenterY
			X 50%
			Y 50%
			W 50%
            H 50%
		`,
		Value: "TEST",
		Events: ui.Events{
			OnMouseButtonUp: func(button uint8, x int32, y int32) bool {
				return false
			},
		},
	})

    s.TemplateWindow.AdoptChild(elButton)

	go s.Loop()

	return
}

// Close our Template state.
func (s *Template) Close() {
	s.TemplateWindow.Destroy()
}

// Loop handles our various state channels.
func (s *Template) Loop() {
	for {
		select {
		case cmd := <-s.Client.CmdChan:
			s.Client.Log.Printf("%v\n", cmd)
		case <-s.Client.ClosedChan:
			s.Client.StateChannel <- client.StateMessage{State: &List{}, Args: nil}
			return
		case <-s.CloseChan:
			return
		}
	}
}