package States

import (
	"github.com/chimera-rpg/go-client/Client"
	"github.com/chimera-rpg/go-client/UI"
	"github.com/chimera-rpg/go-common/Net"
)

type Login struct {
	Client.State
	LoginWindow UI.Window
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

	s.Client.Log.Print("Login State!")

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

	var el_buttons, el_username, el_password, el_confirm, el_email, el_login, el_register, el_previous UI.ElementI

	el_username = UI.NewInputElement(UI.InputElementConfig{
		Style: `
			Origin CenterX CenterY
			X 50%
			Y 10%
			H 20%
			W 100%
			MaxW 200
			MaxH 30
			MinH 25
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
			H 20%
			W 100%
			MaxW 200
			MaxH 30
			MinH 25
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

	el_buttons, _ = UI.NewWindow(UI.WindowConfig{
		Style: `
			X 50%
			Y 80%
			W 60%
			H 30%
			Origin CenterX CenterY
			BackgroundColor 139 139 139 255
		`,
		Parent: &s.LoginWindow,
		RenderFunc: func(w *UI.Window) {
			w.Context.Renderer.Clear()
		},
	})

	el_previous = UI.NewButtonElement(UI.ButtonElementConfig{
		Style: `
			Origin Bottom
			Margin 2%
			H 20%
			W 40%
			MaxH 20
			MaxW 200
			MinW 100
			Padding 6
		`,
		Value: "BACK",
	})

	el_login = UI.NewButtonElement(UI.ButtonElementConfig{
		Style: `
			Origin Right Bottom
			Margin 2%
			H 20%
			W 40%
			MaxH 20
			MaxW 200
			MinW 100
			Padding 6
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
	el_register = UI.NewButtonElement(UI.ButtonElementConfig{
		Style: `
			Origin CenterX CenterY
			X 50%
			Y 80%
			H 20%
			W 100%
			MaxH 20
			MaxW 200
			MinH 25
		`,
		Value: "REGISTER",
		Events: UI.Events{
			OnMouseButtonUp: func(button uint8, x int32, y int32) bool {
				s.Client.Send(Net.Command(Net.CommandLogin{
					Type:  Net.REGISTER,
					User:  el_username.GetValue(),
					Pass:  el_password.GetValue(),
					Email: el_email.GetValue(),
				}))
				return false
			},
		},
	})

	s.Client.Log.Print("Login State")

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
		el_buttons.AdoptChild(el_register)
		el_previous.SetValue("BACK")
		s.LoginWindow.AdoptChild(el_previous)
	}

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
			s.Client.Log.Printf("Server rejected us: %s\n", t.String)
		} else if t.Type == Net.OK {
			s.Client.Log.Printf("Server accepted us: %s\n", t.String)
			s.Client.StateChannel <- Client.StateMessage{&CharacterCreation{}, nil}
			return true
		}
	default:
		s.Client.Log.Print("Server sent non CommandBasic")
		s.Client.StateChannel <- Client.StateMessage{&List{}, nil}
		return true
	}
	return false
}
