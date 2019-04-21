package States

import (
	"fmt"
	"github.com/chimera-rpg/go-client/Client"
	"github.com/chimera-rpg/go-client/UI"
	"github.com/chimera-rpg/go-common/Net"
)

type Login struct {
	Client.State
	LoginWindow UI.Window
	OutputText  UI.ElementI
}

type LoginStateID int

const (
	DefaultState LoginStateID = iota
	RegisterState
	ResetState
)

type LoginState struct {
	state    LoginStateID
	username string
	password string
	email    string
}

func (s *Login) Init(v interface{}) (next Client.StateI, nextArgs interface{}, err error) {
	lstate := LoginState{DefaultState, "", "", ""}

	switch t := v.(type) {
	case LoginState:
		lstate = t
	}

	err = s.LoginWindow.Setup(UI.WindowConfig{
		Value: "Selection",
		Style: `
			W 100%
			H 100%
			BackgroundColor 139 186 139 255
		`,
		Parent: s.Client.RootWindow,
	})

	var el_username, el_password, el_confirm, el_email, el_login, el_previous UI.ElementI

	el_username = UI.NewInputElement(UI.InputElementConfig{
		Style: `
			Origin CenterX CenterY
			X 50%
			Y 10%
			W 100%
			MaxW 200
		`,
		Placeholder: "username",
		Value:       lstate.username,
	})
	el_username.Focus()
	el_email = UI.NewInputElement(UI.InputElementConfig{
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

	el_password = UI.NewInputElement(UI.InputElementConfig{
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
		Events: UI.Events{
			OnKeyDown: func(char uint8, modifiers uint16) bool {
				if char == 13 { // Enter
					el_login.OnMouseButtonUp(1, 0, 0)
				}
				return true
			},
		},
	})
	el_confirm = UI.NewInputElement(UI.InputElementConfig{
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
		Events: UI.Events{
			OnKeyDown: func(char uint8, modifiers uint16) bool {
				if char == 13 { // Enter
					el_login.OnMouseButtonUp(1, 0, 0)
				}
				return true
			},
		},
	})

	el_previous = UI.NewButtonElement(UI.ButtonElementConfig{
		Style: `
			Origin Bottom
			Y 30
			Margin 5%
			W 40%
			MinW 100
		`,
		Value: "BACK",
	})

	el_login = UI.NewButtonElement(UI.ButtonElementConfig{
		Style: `
			Origin Right Bottom
			Y 30
			Margin 5%
			W 40%
			MinW 100
		`,
		Value: "LOGIN",
		Events: UI.Events{
			OnMouseButtonUp: func(button uint8, x int32, y int32) bool {
				s.Client.Send(Net.Command(Net.CommandLogin{
					Type: Net.LOGIN,
					User: el_username.GetValue(),
					Pass: el_password.GetValue(),
				}))
				return false
			},
		},
	})

	s.OutputText = UI.NewTextElement(UI.TextElementConfig{
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
	case DefaultState:
		s.LoginWindow.AdoptChild(el_username)
		s.LoginWindow.AdoptChild(el_password)
		el_previous.SetValue("DISCONNECT")
		el_previous.SetEvents(UI.Events{
			OnMouseButtonUp: func(button uint8, x int32, y int32) bool {
				s.Client.Close()
				return false
			},
		})
		s.LoginWindow.AdoptChild(el_previous)
		s.LoginWindow.AdoptChild(el_login)
	case RegisterState:
		s.LoginWindow.AdoptChild(el_username)
		s.LoginWindow.AdoptChild(el_password)
		s.LoginWindow.AdoptChild(el_confirm)
		s.LoginWindow.AdoptChild(el_email)
		s.LoginWindow.AdoptChild(el_login)
		el_login.SetValue("REGISTER")
		el_login.SetEvents(UI.Events{
			OnMouseButtonUp: func(button uint8, x int32, y int32) bool {
				s.Client.Send(Net.Command(Net.CommandLogin{
					Type:  Net.REGISTER,
					User:  el_username.GetValue(),
					Pass:  el_password.GetValue(),
					Email: el_email.GetValue(),
				}))
				return false
			},
		})
		el_previous.SetValue("BACK")
		s.LoginWindow.AdoptChild(el_previous)
	}
	s.LoginWindow.AdoptChild(s.OutputText)

	go s.Loop()

	return
}

func (s *Login) Close() {
	s.LoginWindow.Destroy()
}

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
			s.Client.StateChannel <- Client.StateMessage{&List{}, nil}
			return
		case <-s.CloseChan:
			return
		}
	}
}

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
			s.Client.StateChannel <- Client.StateMessage{&CharacterCreation{}, msg}
			return true
		}
	default:
		msg := fmt.Sprintf("Server sent non CommandBasic")
		s.Client.Log.Print(msg)
		s.Client.StateChannel <- Client.StateMessage{&List{}, msg}
		return true
	}
	return false
}
