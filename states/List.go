package states

import (
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/ui"
)

// List is the state for showing a server list.
type List struct {
	client.State
	ServersWindow ui.Window
}

// Init our state.
func (s *List) Init(v interface{}) (state client.StateI, nextArgs interface{}, err error) {
	err = s.ServersWindow.Setup(ui.WindowConfig{
		Value: "Server List",
		Style: `
			W 100%
			H 100%
		`,
		Parent: s.Client.RootWindow,
	})

	el := ui.NewTextElement(ui.TextElementConfig{
		Style: `
			PaddingLeft 5%
			PaddingRight 5%
			PaddingTop 5%
			PaddingBottom 5%
			Origin CenterX CenterY
			ContentOrigin CenterX CenterY
			X 50%
			Y 10%
		`,
		Value: "Please choose a server.",
	})

	var elHost, elConnect, elOutputText ui.ElementI

	elHost = ui.NewInputElement(ui.InputElementConfig{
		Style: `
			Origin Bottom
			X 0%
			Y 30
			Margin 5%
			W 60%
		`,
		Placeholder: "host:port",
		Events: ui.Events{
			OnKeyDown: func(char uint8, modifiers uint16) bool {
				if char == 13 { // Enter
					elConnect.OnMouseButtonUp(1, 0, 0)
				}
				return true
			},
		},
	})
	elHost.Focus()
	elConnect = ui.NewButtonElement(ui.ButtonElementConfig{
		Style: `
			Origin Bottom
			X 65%
			Y 30
			Margin 5%
			W 25%
		`,
		Value: "CONNECT",
		Events: ui.Events{
			OnMouseButtonUp: func(which uint8, x int32, y int32) bool {
				s.Client.StateChannel <- client.StateMessage{State: &Handshake{}, Args: elHost.GetValue()}
				return false
			},
		},
	})

	var inString string
	switch t := v.(type) {
	case string:
		inString = t
	case error:
		inString = t.Error()
	default:
		inString = "Type in an address or select a server from above and connect."
	}

	elOutputText = ui.NewTextElement(ui.TextElementConfig{
		Style: `
			Origin CenterX Bottom
			ContentOrigin CenterX CenterY
			ForegroundColor 255 255 255 128
			BackgroundColor 0 0 0 128
			Y 0
			X 50%
			W 100%
		`,
		Value: inString,
	})

	imageData, err := s.Client.DataManager.GetBytes(s.Client.DataManager.GetDataPath("ui/loading.png"))

	elImg := ui.NewImageElement(ui.ImageElementConfig{
		Style: `
			X 50%
			Y 50%
			W 48
			H 48
			Origin CenterX CenterY
		`,
		Image: imageData,
	})

	s.ServersWindow.AdoptChild(el)
	s.ServersWindow.AdoptChild(elHost)
	s.ServersWindow.AdoptChild(elConnect)
	s.ServersWindow.AdoptChild(elOutputText)
	el.AdoptChild(elImg)

	elTest := ui.NewTextElement(ui.TextElementConfig{
		Style: `
			X 50%
			Y 50%
			Origin CenterX CenterY
		`,
		Value: "Test",
	})
	elImg.AdoptChild(elTest)

	return
}

// Close our state.
func (s *List) Close() {
	s.ServersWindow.Destroy()
}
