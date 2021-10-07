package list

import (
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/ui"
)

// List is the state for showing a server list.
type List struct {
	client.State
	ServersContainer ui.Container
}

// Init our state.
func (s *List) Init(v interface{}) (state client.StateI, nextArgs interface{}, err error) {
	err = s.ServersContainer.Setup(ui.ContainerConfig{
		Value: "Server List",
		Style: `
			W 100%
			H 100%
		`,
	})
	s.Client.RootWindow.AdoptChannel <- s.ServersContainer.This

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
		Value:         s.Client.DataManager.Config.LastServer,
		Placeholder:   "host:port",
		SubmitOnEnter: true,
		Events: ui.Events{
			OnTextSubmit: func(str string) bool {
				elConnect.OnPressed(1, 0, 0)
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
			OnPressed: func(which uint8, x int32, y int32) bool {
				s.Client.DataManager.Config.LastServer = elHost.GetValue()
				s.Client.CurrentServer = elHost.GetValue()
				s.Client.StateChannel <- client.StateMessage{Push: true, State: &Handshake{}, Args: elHost.GetValue()}
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

	imageData, err := s.Client.DataManager.GetImage(s.Client.DataManager.GetDataPath("ui/loading.png"))

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

	s.ServersContainer.AdoptChannel <- el
	s.ServersContainer.AdoptChannel <- elHost
	s.ServersContainer.AdoptChannel <- elConnect
	s.ServersContainer.AdoptChannel <- elOutputText
	el.GetAdoptChannel() <- elImg

	elTest := ui.NewTextElement(ui.TextElementConfig{
		Style: `
			X 50%
			Y 50%
			Origin CenterX CenterY
		`,
		Value: "Test",
	})
	elTest.GetAdoptChannel() <- elImg

	return
}

// Close our state.
func (s *List) Close() {
	s.ServersContainer.GetDestroyChannel() <- true
}

func (s *List) Leave() {
	s.ServersContainer.GetUpdateChannel() <- ui.UpdateHidden(true)
}

func (s *List) Enter(args ...interface{}) {
	s.ServersContainer.GetUpdateChannel() <- ui.UpdateHidden(false)
}
