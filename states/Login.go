package states

import (
	"fmt"

	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-common/Net"
)

// Login is the state responsible for logging in, registering an account,
// or recovering an account.
type Login struct {
	client.State
	LoginWindow ui.Window
	OutputText  ui.ElementI
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
	email    string
}

// Init our Login state.
func (s *Login) Init(v interface{}) (next client.StateI, nextArgs interface{}, err error) {
	lstate := LoginState{defaultState, "", "", ""}

	switch t := v.(type) {
	case LoginState:
		lstate = t
	}

	err = s.LoginWindow.Setup(ui.WindowConfig{
		Value: "Selection",
		Style: `
			W 100%
			H 100%
			BackgroundColor 139 186 139 255
		`,
		Parent: s.Client.RootWindow,
	})

	var elUsername, elPassword, elConfirm, elEmail, elLogin, elPrevious ui.ElementI

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
			OnKeyDown: func(char uint8, modifiers uint16) bool {
				if char == 13 { // Enter
					elLogin.OnMouseButtonUp(1, 0, 0)
				}
				return true
			},
		},
	})
	elUsername.Focus()
	elEmail = ui.NewInputElement(ui.InputElementConfig{
		Style: `
			Origin CenterX CenterY
			X 60%
			Y 10%
			W 100%
			MaxW 200
		`,
		Placeholder: "email",
		Value:       lstate.email,
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
		Value:       lstate.username,
		Events: ui.Events{
			OnKeyDown: func(char uint8, modifiers uint16) bool {
				if char == 13 { // Enter
					elLogin.OnMouseButtonUp(1, 0, 0)
				}
				return true
			},
		},
	})
	elConfirm = ui.NewInputElement(ui.InputElementConfig{
		Style: `
			Origin CenterX CenterY
			X 50%
			Y 60%
			H 20%
			W 100%
			MaxW 200
			MaxH 30
			MinH 25
			ForegroundColor 255 0 0 255
		`,
		Password:    true,
		Placeholder: "password confirm",
		Events: ui.Events{
			OnKeyDown: func(char uint8, modifiers uint16) bool {
				if char == 13 { // Enter
					elLogin.OnMouseButtonUp(1, 0, 0)
				}
				return true
			},
		},
	})

	elPrevious = ui.NewButtonElement(ui.ButtonElementConfig{
		Style: `
			Origin Bottom
			Y 30
			Margin 5%
			W 40%
			MinW 100
		`,
		Value: "BACK",
	})

	elLogin = ui.NewButtonElement(ui.ButtonElementConfig{
		Style: `
			Origin Right Bottom
			Y 30
			Margin 5%
			W 40%
			MinW 100
		`,
		Value: "LOGIN",
		Events: ui.Events{
			OnMouseButtonUp: func(button uint8, x int32, y int32) bool {
				s.Client.Send(Net.Command(Net.CommandLogin{
					Type: Net.LOGIN,
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
			ForegroundColor 255 255 255 128
			BackgroundColor 0 0 0 128
			Y 0
			X 50%
			W 100%
		`,
		Value: "Connected.",
	})

	switch lstate.state {
	case defaultState:
		s.LoginWindow.AdoptChild(elUsername)
		s.LoginWindow.AdoptChild(elPassword)
		elPrevious.SetValue("DISCONNECT")
		elPrevious.SetEvents(ui.Events{
			OnMouseButtonUp: func(button uint8, x int32, y int32) bool {
				s.Client.Close()
				return false
			},
		})
		s.LoginWindow.AdoptChild(elPrevious)
		s.LoginWindow.AdoptChild(elLogin)
	case registerState:
		s.LoginWindow.AdoptChild(elUsername)
		s.LoginWindow.AdoptChild(elPassword)
		s.LoginWindow.AdoptChild(elConfirm)
		s.LoginWindow.AdoptChild(elEmail)
		s.LoginWindow.AdoptChild(elLogin)
		elLogin.SetValue("REGISTER")
		elLogin.SetEvents(ui.Events{
			OnMouseButtonUp: func(button uint8, x int32, y int32) bool {
				s.Client.Send(Net.Command(Net.CommandLogin{
					Type:  Net.REGISTER,
					User:  elUsername.GetValue(),
					Pass:  elPassword.GetValue(),
					Email: elEmail.GetValue(),
				}))
				return false
			},
		})
		elPrevious.SetValue("BACK")
		s.LoginWindow.AdoptChild(elPrevious)
	}
	s.LoginWindow.AdoptChild(s.OutputText)

	go s.Loop()

	return
}

// Close our Login state.
func (s *Login) Close() {
	s.LoginWindow.Destroy()
}

// Loop handles our various state channels.
func (s *Login) Loop() {
	for {
		select {
		case cmd := <-s.Client.CmdChan:
			ret := s.HandleNet(cmd)
			if ret {
				//return
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
func (s *Login) HandleNet(cmd Net.Command) bool {
	switch t := cmd.(type) {
	case Net.CommandBasic:
		s.Client.Log.Print("Got basic")
		if t.Type == Net.REJECT {
			msg := fmt.Sprintf("Server rejected us: %s\n", t.String)
			s.OutputText.SetValue(msg)
			s.Client.Log.Printf(msg)
		} else if t.Type == Net.OK {
			msg := fmt.Sprintf("Server accepted us: %s\n", t.String)
			s.OutputText.SetValue(msg)
			s.Client.Log.Printf(msg)
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
