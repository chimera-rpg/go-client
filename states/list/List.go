package list

import (
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/ui"
)

// List is the state for showing a server list.
type List struct {
	client.State
	//ServersContainer ui.Container
	ServersContainer ui.ElementI
	layout           ui.LayoutEntry
}

// Init our state.
func (s *List) Init(v interface{}) (state client.StateI, nextArgs interface{}, err error) {
	imageData, err := s.Client.DataManager.GetImage(s.Client.DataManager.GetDataPath("ui/loading.png"))
	if err != nil {
		panic(err)
	}

	var inString string
	switch t := v.(type) {
	case string:
		inString = t
	case error:
		inString = t.Error()
	default:
		inString = "Type in an address or select a server from above and connect."
	}

	s.layout = s.Client.DataManager.Layouts["Servers"].Children[0].Generate(s.Client.DataManager.Styles["Servers"], map[string]interface{}{
		"List": ui.ContainerConfig{
			Value: "Server List",
		},
		"ChooseServer": ui.TextElementConfig{
			Value: "Please choose a server.",
		},
		"HostInput": ui.InputElementConfig{
			Value:         s.Client.DataManager.Config.LastServer,
			Placeholder:   "host:port",
			SubmitOnEnter: true,
			Events: ui.Events{
				OnTextSubmit: func(str string) bool {
					s.layout.Find("ConnectButton").Element.OnPressed(1, 0, 0)
					return true
				},
			},
		},
		"SplashImage": ui.ImageElementConfig{
			Image: imageData,
		},
		"ConnectButton": ui.ButtonElementConfig{
			Value: "CONNECT",
			Events: ui.Events{
				OnPressed: func(which uint8, x int32, y int32) bool {
					elHost := s.layout.Find("HostInput").Element
					s.Client.DataManager.Config.LastServer = elHost.GetValue()
					s.Client.CurrentServer = elHost.GetValue()
					s.Client.StateChannel <- client.StateMessage{Push: true, State: &Handshake{}, Args: elHost.GetValue()}
					return false
				},
			},
		},
		"OutputText": ui.TextElementConfig{
			Value: inString,
		},
	})

	s.Client.RootWindow.AdoptChannel <- s.layout.Find("List").Element
	s.layout.Find("HostInput").Element.Focus()

	return
}

// Close our state.
func (s *List) Close() {
	s.layout.Find("List").Element.GetDestroyChannel() <- true
}

func (s *List) Leave() {
	s.layout.Find("List").Element.GetUpdateChannel() <- ui.UpdateHidden(true)
}

func (s *List) Enter(args ...interface{}) {
	s.layout.Find("List").Element.GetUpdateChannel() <- ui.UpdateHidden(false)
}
