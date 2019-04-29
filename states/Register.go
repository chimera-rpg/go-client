package states

import (
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/ui"
)

// Register is a state for user Registration.
// It follows after Login and returns to Login.
type Register struct {
	client.State
	RegisterWindow ui.Window
}

// Init our Register state.
func (s *Register) Init(v interface{}) (next client.StateI, nextArgs interface{}, err error) {
	var elUsername, elPassword, elPassword2, elBack, elConfirm ui.ElementI

	err = s.RegisterWindow.Setup(ui.WindowConfig{
		Value: "Register",
		Style: `
			W 100%
			H 100%
			BackgroundColor 139 186 139 255
		`,
		Parent: s.Client.RootWindow,
	})

	elUsername = ui.NewInputElement(ui.InputElementConfig{
		Style: `
			Origin CenterX CenterY
			X 50%
			Y 10%
			W 100%
			MaxW 200
		`,
		Placeholder: "username",
		Events: ui.Events{
			OnAdopted: func(parent ui.ElementI) {
				elUsername.Focus()
			},
			OnKeyDown: func(char uint8, modifiers uint16) bool {
				if char == 13 { // Enter
					//elLogin.OnMouseButtonUp(1, 0, 0)
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
		Events: ui.Events{
			OnKeyDown: func(char uint8, modifiers uint16) bool {
				if char == 13 { // Enter
					//elLogin.OnMouseButtonUp(1, 0, 0)
				}
				return true
			},
		},
	})
	elPassword2 = ui.NewInputElement(ui.InputElementConfig{
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
		Placeholder: "confirm password",
		Events: ui.Events{
			OnKeyDown: func(char uint8, modifiers uint16) bool {
				if char == 13 { // Enter
					//elLogin.OnMouseButtonUp(1, 0, 0)
				}
				return true
			},
		},
	})

	elConfirm = ui.NewButtonElement(ui.ButtonElementConfig{
		Style: `
			Origin CenterX CenterY
			X 50%
			Y 70%
			H 20%
			W 100%
			MaxW 200
			MaxH 30
			MinH 25
		`,
		Value: "Confirm",
		Events: ui.Events{
			OnMouseButtonUp: func(button uint8, x, y int32) bool {
				return true
			},
		},
	})

	elBack = ui.NewButtonElement(ui.ButtonElementConfig{
		Style: `
			Origin Bottom
			Y 30
			Margin 5%
			W 40%
			MinW 100
		`,
		Value: "Back",
		Events: ui.Events{
			OnMouseButtonUp: func(button uint8, x, y int32) bool {
				s.Client.StateChannel <- client.StateMessage{State: &Login{}}
				return true
			},
		},
	})

	s.RegisterWindow.AdoptChild(elUsername)
	s.RegisterWindow.AdoptChild(elPassword)
	s.RegisterWindow.AdoptChild(elPassword2)
	//s.RegisterWindow.AdoptChild(elEmail)
	s.RegisterWindow.AdoptChild(elConfirm)
	s.RegisterWindow.AdoptChild(elBack)
	//s.RegisterWindow.AdoptChild(elText)

	go s.Loop()

	return
}

// Close our Register state.
func (s *Register) Close() {
	s.RegisterWindow.Destroy()
}

// Loop handles our various state channels.
func (s *Register) Loop() {
	for {
		select {
		case cmd := <-s.Client.CmdChan:
			s.Client.Log.Printf("%v\n", cmd)
		case <-s.Client.ClosedChan:
			s.Client.StateChannel <- client.StateMessage{State: &List{}, Args: nil}
			return
		case <-s.CloseChan:
			return
		}
	}
}
