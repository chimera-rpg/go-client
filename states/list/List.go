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
		Style: s.Client.DataManager.Styles["Servers"]["List"],
	})
	s.Client.RootWindow.AdoptChannel <- s.ServersContainer.This

	el := ui.NewTextElement(ui.TextElementConfig{
		Style: s.Client.DataManager.Styles["Servers"]["ChooseServer"],
		Value: "Please choose a server.",
	})

	var elHost, elConnect, elOutputText ui.ElementI

	elHost = ui.NewInputElement(ui.InputElementConfig{
		Style:         s.Client.DataManager.Styles["Servers"]["HostInput"],
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
		Style: s.Client.DataManager.Styles["Servers"]["ConnectButton"],
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
		Style: s.Client.DataManager.Styles["Servers"]["OutputText"],
		Value: inString,
	})

	imageData, err := s.Client.DataManager.GetImage(s.Client.DataManager.GetDataPath("ui/loading.png"))

	elImg := ui.NewImageElement(ui.ImageElementConfig{
		Style: s.Client.DataManager.Styles["Servers"]["SplashImage"],
		Image: imageData,
	})

	s.ServersContainer.AdoptChannel <- el
	s.ServersContainer.AdoptChannel <- elHost
	s.ServersContainer.AdoptChannel <- elConnect
	s.ServersContainer.AdoptChannel <- elOutputText
	el.GetAdoptChannel() <- elImg

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
