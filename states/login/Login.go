package login

import (
	"flag"
	"fmt"

	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/config"
	"github.com/chimera-rpg/go-client/states/game"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-server/network"
)

// Login is the state responsible for logging in, registering an account,
// or recovering an account.
type Login struct {
	client.State
	layout           ui.LayoutEntry
	rememberPassword bool
	pendingLogin     bool
}

// StateID represents the current sub state of the Login state.
type LoginStateID int

const (
	defaultState LoginStateID = iota
	registerState
	resetState
)

// LoginState is our Login state's current... state. Fancy that.
type LoginState struct {
	state    LoginStateID
	username string
	password string
	message  string
}

// Init our Login state.
func (s *Login) Init(v interface{}) (next client.StateI, nextArgs interface{}, err error) {
	lstate := LoginState{
		state:    defaultState,
		username: "",
		password: "",
		message:  "Connected.",
	}

	if v, ok := s.Client.DataManager.Config.Servers[s.Client.CurrentServer]; ok {
		lstate.username = v.Username
		lstate.password = v.Password
		s.rememberPassword = v.RememberPassword
	}

	switch t := v.(type) {
	case LoginState:
		lstate = t
	}

	remember := "remember: no"
	if s.rememberPassword {
		remember = "remember: yes"
	}

	s.layout = s.Client.DataManager.Layouts["Login"][0].Generate(s.Client.DataManager.Styles["Login"], map[string]interface{}{
		"Container": ui.ContainerConfig{
			Value: "Selection",
		},
		"UsernameInput": ui.InputElementConfig{
			Placeholder: "username",
			Value:       lstate.username,
			Events: ui.Events{
				OnAdopted: func(parent ui.ElementI) {
					s.layout.Find("UsernameInput").Element.Focus()
				},
				OnKeyDown: func(char uint8, modifiers uint16, repeat bool) bool {
					if char == 13 { // Enter
						s.layout.Find("LoginButton").Element.OnPressed(1, 0, 0)
					}
					return true
				},
			},
		},
		"PasswordInput": ui.InputElementConfig{
			Password:    true,
			Placeholder: "password",
			Value:       lstate.password,
			Events: ui.Events{
				OnKeyDown: func(char uint8, modifiers uint16, repeat bool) bool {
					if char == 13 { // Enter
						s.layout.Find("LoginButton").Element.OnPressed(1, 0, 0)
					}
					return true
				},
			},
		},
		"RememberButton": ui.ButtonElementConfig{
			Value: remember,
			Events: ui.Events{
				OnPressed: func(button uint8, x, y int32) bool {
					el := s.layout.Find("RememberButton").Element
					s.rememberPassword = !s.rememberPassword
					if s.rememberPassword {
						el.GetUpdateChannel() <- ui.UpdateValue{Value: "remember: yes"}
					} else {
						el.GetUpdateChannel() <- ui.UpdateValue{Value: "remember: no"}
					}
					return true
				},
			},
		},
		"DisconnectButton": ui.ButtonElementConfig{
			Value: "DISCONNECT",
			Events: ui.Events{
				OnPressed: func(button uint8, x int32, y int32) bool {
					s.Client.Close()
					return false
				},
			},
		},
		"RegisterButton": ui.ButtonElementConfig{
			Value: "REGISTER",
			Events: ui.Events{
				OnPressed: func(button uint8, x int32, y int32) bool {
					s.Client.StateChannel <- client.StateMessage{State: &Register{}, Args: nil}
					return false
				},
			},
		},
		"LoginButton": ui.ButtonElementConfig{
			Style: s.Client.DataManager.Styles["Login"]["LoginButton"],
			Value: "LOGIN",
			Events: ui.Events{
				OnPressed: func(button uint8, x int32, y int32) bool {
					if !s.pendingLogin {
						s.pendingLogin = true
						s.Client.Send(network.Command(network.CommandLogin{
							Type: network.Login,
							User: s.layout.Find("UsernameInput").Element.GetValue(),
							Pass: s.layout.Find("PasswordInput").Element.GetValue(),
						}))
					}
					return false
				},
			},
		},
		"OutputText": ui.TextElementConfig{
			Value: lstate.message,
		},
	})

	s.Client.RootWindow.AdoptChannel <- s.layout.Find("Container").Element

	go s.Loop()

	return
}

func (s *Login) Leave() {
	s.layout.Find("Container").Element.GetUpdateChannel() <- ui.UpdateHidden(true)
}

func (s *Login) Enter(args ...interface{}) {
	s.layout.Find("Container").Element.GetUpdateChannel() <- ui.UpdateHidden(false)
}

// Close our Login state.
func (s *Login) Close() {
	s.layout.Find("Container").Element.GetDestroyChannel() <- true
}

// Loop handles our various state channels.
func (s *Login) Loop() {
	// Attempt to automatically log in if username and password have been provided.
	username := flag.Lookup("username")
	password := flag.Lookup("password")
	if username.Value.String() != username.DefValue && password.Value.String() != username.DefValue {
		s.Client.Send(network.Command(network.CommandLogin{
			Type: network.Login,
			User: username.Value.String(),
			Pass: password.Value.String(),
		}))
	}
	for {
		if !s.Running {
			continue
		}
		select {
		case cmd := <-s.Client.CmdChan:
			ret := s.HandleNet(cmd)
			if ret {
				return
			}
		case <-s.Client.ClosedChan:
			s.Client.Log.Print("Lost connection to server.")
			s.Client.StateChannel <- client.StateMessage{PopToTop: true, Args: nil}
			return
		case <-s.CloseChan:
			return
		}
	}
}

// HandleNet handles the network commands received in Loop().
func (s *Login) HandleNet(cmd network.Command) bool {
	switch t := cmd.(type) {
	case network.CommandRejoin: // If we are sent a rejoin command, just immediately head over to game state.
		s.Client.StateChannel <- client.StateMessage{Push: true, State: &game.Game{}, Args: nil}
		return true
	case network.CommandBasic:
		s.Client.Log.Print("Got basic")
		if t.Type == network.Reject {
			msg := fmt.Sprintf("Server rejected us: %s", t.String)
			s.layout.Find("OutputText").Element.GetUpdateChannel() <- ui.UpdateValue{Value: msg}
			s.Client.Log.Println(msg)
			s.pendingLogin = false
		} else if t.Type == network.Okay {
			msg := fmt.Sprintf("Server accepted us: %s", t.String)
			s.layout.Find("OutputText").Element.GetUpdateChannel() <- ui.UpdateValue{Value: msg}
			s.Client.Log.Println(msg)
			// Set our username and password for this server.
			serverName := s.Client.CurrentServer
			if s.Client.DataManager.Config.Servers == nil {
				s.Client.DataManager.Config.Servers = make(map[string]*config.ServerConfig)
			}
			if _, ok := s.Client.DataManager.Config.Servers[serverName]; !ok {
				s.Client.DataManager.Config.Servers[serverName] = &config.ServerConfig{}
			}
			s.Client.DataManager.Config.Servers[serverName].Username = s.layout.Find("UsernameInput").Element.GetValue()
			if s.rememberPassword {
				s.Client.DataManager.Config.Servers[serverName].Password = s.layout.Find("PasswordInput").Element.GetValue()
			} else {
				s.Client.DataManager.Config.Servers[serverName].Password = ""
			}
			s.Client.DataManager.Config.Servers[serverName].RememberPassword = s.rememberPassword
			s.Client.StateChannel <- client.StateMessage{Push: true, State: &CharacterSelection{}, Args: msg}
			return true
		}
	default:
		msg := fmt.Sprintf("Server sent non CommandBasic %d", t.GetType())
		s.Client.Log.Print(msg)
		s.Client.StateChannel <- client.StateMessage{PopToTop: true, Args: msg}
		return true
	}
	return false
}
