package states

import (
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/ui"
)

type List struct {
	client.State
	ServersWindow ui.Window
}

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

	var el_host, el_connect, el_output_text ui.ElementI

	el_host = ui.NewInputElement(ui.InputElementConfig{
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
					el_connect.OnMouseButtonUp(1, 0, 0)
				}
				return true
			},
		},
	})
	el_host.Focus()
	el_connect = ui.NewButtonElement(ui.ButtonElementConfig{
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
				s.Client.StateChannel <- client.StateMessage{&Handshake{}, el_host.GetValue()}
				return false
			},
		},
	})

	var in_string string
	switch t := v.(type) {
	case string:
		in_string = t
	case error:
		in_string = t.Error()
	default:
		in_string = "Type in an address or select a server from above and connect."
	}

	el_output_text = ui.NewTextElement(ui.TextElementConfig{
		Style: `
			Origin CenterX Bottom
			ContentOrigin CenterX CenterY
			ForegroundColor 255 255 255 128
			BackgroundColor 0 0 0 128
			Y 0
			X 50%
			W 100%
		`,
		Value: in_string,
	})

	el_img := ui.NewImageElement(ui.ImageElementConfig{
		Style: `
			X 50%
			Y 50%
			W 48
			H 48
			Origin CenterX CenterY
		`,
		Image: s.Client.GetPNGData("ui/loading.png"),
	})

	s.ServersWindow.AdoptChild(el)
	s.ServersWindow.AdoptChild(el_host)
	s.ServersWindow.AdoptChild(el_connect)
	s.ServersWindow.AdoptChild(el_output_text)
	el.AdoptChild(el_img)

	el_test := ui.NewTextElement(ui.TextElementConfig{
		Style: `
			X 50%
			Y 50%
			Origin CenterX CenterY
		`,
		Value: "Test",
	})
	el_img.AdoptChild(el_test)

	return
}

func (s *List) Close() {
	s.ServersWindow.Destroy()
}
