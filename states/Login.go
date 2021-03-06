package states

import (
	"flag"
	"fmt"

	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-common/network"
)

// Login is the state responsible for logging in, registering an account,
// or recovering an account.
type Login struct {
	client.State
	LoginContainer ui.Container
	OutputText     ui.ElementI
}

// LoginStateID represents the current sub state of the Login state.
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

	switch t := v.(type) {
	case LoginState:
		lstate = t
	}

	err = s.LoginContainer.Setup(ui.ContainerConfig{
		Value: "Selection",
		Style: `
			W 100%
			H 100%
			BackgroundColor 139 186 139 255
		`,
	})

	var elUsername, elPassword, elLogin, elRegister, elDisconnect ui.ElementI

	elUsername = ui.NewInputElement(ui.InputElementConfig{
		Style: `
			Origin CenterX CenterY
			X 50%
			Y 10%
			W 100%
			MaxW 200
		`,
		Placeholder: "username",
		Value:       lstate.username,
		Events: ui.Events{
			OnAdopted: func(parent ui.ElementI) {
				elUsername.Focus()
			},
			OnKeyDown: func(char uint8, modifiers uint16, repeat bool) bool {
				if char == 13 { // Enter
					elLogin.OnMouseButtonUp(1, 0, 0)
				}
				return true
			},
		},
	})

	elPassword = ui.NewInputElement(ui.InputElementConfig{
		Style: `
			Origin CenterX CenterY
			X 50%
			Y 40%
			H 20%
			W 100%
			MaxW 200
			MaxH 30
			MinH 25
			ForegroundColor 255 0 0 255
		`,
		Password:    true,
		Placeholder: "password",
		Value:       lstate.password,
		Events: ui.Events{
			OnKeyDown: func(char uint8, modifiers uint16, repeat bool) bool {
				if char == 13 { // Enter
					elLogin.OnMouseButtonUp(1, 0, 0)
				}
				return true
			},
		},
	})

	elDisconnect = ui.NewButtonElement(ui.ButtonElementConfig{
		Style: `
			Origin Bottom
			Y 30
			Margin 5%
			W 40%
			MinW 100
		`,
		Value: "DISCONNECT",
		Events: ui.Events{
			OnMouseButtonUp: func(button uint8, x int32, y int32) bool {
				s.Client.Close()
				return false
			},
		},
	})

	elRegister = ui.NewButtonElement(ui.ButtonElementConfig{
		Style: `
			Origin Bottom
			X 50%
			Y 30
			Margin 5%
			W 40%
			MinW 100
		`,
		Value: "REGISTER",
		Events: ui.Events{
			OnMouseButtonUp: func(button uint8, x int32, y int32) bool {
				s.Client.StateChannel <- client.StateMessage{State: &Register{}}
				return false
			},
		},
	})

	elLogin = ui.NewButtonElement(ui.ButtonElementConfig{
		Style: `
			Origin CenterX CenterY
			X 50%
			Y 60%
			W 40%
		`,
		Value: "LOGIN",
		Events: ui.Events{
			OnMouseButtonUp: func(button uint8, x int32, y int32) bool {
				s.Client.Send(network.Command(network.CommandLogin{
					Type: network.Login,
					User: elUsername.GetValue(),
					Pass: elPassword.GetValue(),
				}))
				return false
			},
		},
	})

	s.OutputText = ui.NewTextElement(ui.TextElementConfig{
		Style: `
			Origin CenterX Bottom
			ContentOrigin CenterX CenterY
			ForegroundColor 255 255 255 255
			BackgroundColor 0 0 0 128
			Y 0
			X 50%
			W 100%
		`,
		Value: lstate.message,
	})

	s.LoginContainer.AdoptChannel <- elUsername
	s.LoginContainer.AdoptChannel <- elPassword
	s.LoginContainer.AdoptChannel <- elLogin
	s.LoginContainer.AdoptChannel <- elDisconnect
	s.LoginContainer.AdoptChannel <- elRegister

	s.LoginContainer.AdoptChannel <- s.OutputText

	s.Client.RootWindow.AdoptChannel <- s.LoginContainer.This

	go s.Loop()

	return
}

// Close our Login state.
func (s *Login) Close() {
	s.LoginContainer.GetDestroyChannel() <- true
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
		select {
		case cmd := <-s.Client.CmdChan:
			ret := s.HandleNet(cmd)
			if ret {
				return
			}
		case <-s.Client.ClosedChan:
			s.Client.Log.Print("Lost connection to server.")
			s.Client.StateChannel <- client.StateMessage{State: &List{}, Args: nil}
			return
		case <-s.CloseChan:
			return
		}
	}
}

// HandleNet handles the network commands received in Loop().
func (s *Login) HandleNet(cmd network.Command) bool {
	switch t := cmd.(type) {
	case network.CommandBasic:
		s.Client.Log.Print("Got basic")
		if t.Type == network.Reject {
			msg := fmt.Sprintf("Server rejected us: %s", t.String)
			s.OutputText.GetUpdateChannel() <- ui.UpdateValue{Value: msg}
			s.Client.Log.Println(msg)
		} else if t.Type == network.Okay {
			msg := fmt.Sprintf("Server accepted us: %s", t.String)
			s.OutputText.GetUpdateChannel() <- ui.UpdateValue{Value: msg}
			s.Client.Log.Println(msg)
			s.Client.StateChannel <- client.StateMessage{State: &CharacterCreation{}, Args: msg}
			return true
		}
	default:
		msg := fmt.Sprintf("Server sent non CommandBasic")
		s.Client.Log.Print(msg)
		s.Client.StateChannel <- client.StateMessage{State: &List{}, Args: msg}
		return true
	}
	return false
}
