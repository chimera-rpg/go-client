package States

import (
	"github.com/chimera-rpg/go-client/Client"
	"github.com/chimera-rpg/go-client/UI"
)

type List struct {
	Client.State
	ServersWindow UI.Window
}

func (s *List) Init(v interface{}) (state Client.StateI, nextArgs interface{}, err error) {
	err = s.ServersWindow.Setup(UI.WindowConfig{
		Value: "Server List",
		Style: `
			W 100%
			H 100%
		`,
		Parent: s.Client.RootWindow,
	})

	el := UI.NewTextElement(UI.TextElementConfig{
		Style: `
			PaddingLeft 5%
			PaddingRight 5%
			PaddingTop 5%
			PaddingBottom 5%
			Origin CenterX CenterY
			X 50%
			Y 10%
		`,
		Value: "Please choose a server:",
	})

	var el_host, el_connect, el_output_text UI.ElementI

	el_host = UI.NewInputElement(UI.InputElementConfig{
		Style: `
			X 10%
			Y 80%
			W 60%
			H 30
		`,
		Placeholder: "host:port",
		Events: UI.Events{
			OnKeyDown: func(char uint8, modifiers uint16) bool {
				if char == 13 { // Enter
					el_connect.OnMouseButtonUp(1, 0, 0)
				}
				return true
			},
		},
	})
	el_connect = UI.NewButtonElement(UI.ButtonElementConfig{
		Style: `
			X 80%
			Y 80%
			W 10%
			H 30
		`,
		Value: "Connect",
		Events: UI.Events{
			OnMouseButtonUp: func(which uint8, x int32, y int32) bool {
				s.Client.StateChannel <- Client.StateMessage{&Handshake{}, el_host.GetValue()}
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

	el_output_text = UI.NewTextElement(UI.TextElementConfig{
		Style: `
			Origin CenterX Bottom
			ContentOrigin CenterX CenterY
			ForegroundColor 255 255 255 128
			BackgroundColor 0 0 0 128
			Y 0
			X 50%
			W 100%
			H 30
			Padding 6
		`,
		Value: in_string,
	})

	el_img := UI.NewImageElement(UI.ImageElementConfig{
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

	el_test := UI.NewTextElement(UI.TextElementConfig{
		Style: `
			X 50%
			Y 50%
			Origin CenterX CenterY
		`,
		Value: "Test",
	})
	el_img.AdoptChild(el_test)

	s.Client.Print("Please choose a server: ")
	return
}

func (s *List) Close() {
	s.ServersWindow.Destroy()
}
