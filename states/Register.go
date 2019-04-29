package states

import (
	"regexp"

	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/ui"
)

// Register is a state for user Registration.
// It follows after Login and returns to Login.
type Register struct {
	client.State
	RegisterWindow, InputWindow                  ui.Window
	OutputText                                   ui.ElementI
	elUsername, elPassword, elPassword2, elEmail ui.ElementI
	textUsername, textPassword, textEmail        ui.ElementI
}

// Init our Register state.
func (s *Register) Init(v interface{}) (next client.StateI, nextArgs interface{}, err error) {
	var elBack, elConfirm ui.ElementI

	err = s.RegisterWindow.Setup(ui.WindowConfig{
		Value: "Register",
		Style: `
			W 100%
			H 100%
			BackgroundColor 139 186 139 255
		`,
		Parent: s.Client.RootWindow,
	})

	err = s.InputWindow.Setup(ui.WindowConfig{
		Style: `
			Origin CenterX CenterY
			X 50%
			Y 50%
			MinW 100
			W 100%
			MinH 300
			MaxH 100%
		`,
		Parent: &s.RegisterWindow,
	})

	s.elUsername = ui.NewInputElement(ui.InputElementConfig{
		Style: `
			Origin CenterX
			X 50%
			Y 0
			W 100%
			MaxW 200
			MaxH 30
			MinH 25
		`,
		Placeholder: "username",
		Events: ui.Events{
			OnAdopted: func(parent ui.ElementI) {
				s.elUsername.Focus()
			},
			OnChange: func() bool {
				str, ok := s.verifyUsername()
				if !ok {
					s.textUsername.GetStyle().ForegroundColor = ui.Color{
						R: 200,
						G: 32,
						B: 32,
						A: 255,
					}
				} else {
					s.textUsername.GetStyle().ForegroundColor = ui.Color{
						R: 32,
						G: 200,
						B: 32,
						A: 255,
					}
				}
				s.textUsername.SetValue(str)
				return true
			},
		},
	})
	s.textUsername = ui.NewTextElement(ui.TextElementConfig{
		Style: `
			Origin CenterX
			ContentOrigin CenterX CenterY
			X 50%
			Y 30
			H 20
			W 100%
		`,
	})

	s.elPassword = ui.NewInputElement(ui.InputElementConfig{
		Style: `
			Origin CenterX
			X 50%
			Y 60
			W 100%
			MaxW 200
			MaxH 30
			MinH 25
		`,
		Password:    true,
		Placeholder: "password",
		Events: ui.Events{
			OnChange: func() bool {
				str, ok := s.verifyPassword()
				if !ok {
					s.textPassword.GetStyle().ForegroundColor = ui.Color{
						R: 200,
						G: 32,
						B: 32,
						A: 255,
					}
				} else {
					s.textPassword.GetStyle().ForegroundColor = ui.Color{
						R: 32,
						G: 200,
						B: 32,
						A: 255,
					}
				}
				s.textPassword.SetValue(str)
				return true
			},
		},
	})
	s.elPassword2 = ui.NewInputElement(ui.InputElementConfig{
		Style: `
			Origin CenterX
			X 50%
			Y 110
			W 100%
			MaxW 200
			MaxH 30
			MinH 25
		`,
		Password:    true,
		Placeholder: "confirm password",
		Events: ui.Events{
			OnChange: func() bool {
				str, ok := s.verifyPassword()
				if !ok {
					s.textPassword.GetStyle().ForegroundColor = ui.Color{
						R: 200,
						G: 32,
						B: 32,
						A: 255,
					}
				} else {
					s.textPassword.GetStyle().ForegroundColor = ui.Color{
						R: 32,
						G: 200,
						B: 32,
						A: 255,
					}
				}
				s.textPassword.SetValue(str)
				return true
			},
		},
	})
	s.textPassword = ui.NewTextElement(ui.TextElementConfig{
		Style: `
			Origin CenterX
			ContentOrigin CenterX CenterY
			X 50%
			Y 140
			H 20
			W 100%
		`,
	})

	s.elEmail = ui.NewInputElement(ui.InputElementConfig{
		Style: `
			Origin CenterX
			X 50%
			Y 170
			W 100%
			MaxW 200
			MaxH 30
			MinH 25
		`,
		Placeholder: "email",
		Events: ui.Events{
			OnChange: func() bool {
				str, ok := s.verifyEmail()
				if !ok {
					s.textEmail.GetStyle().ForegroundColor = ui.Color{
						R: 200,
						G: 200,
						B: 32,
						A: 255,
					}
				} else {
					s.textEmail.GetStyle().ForegroundColor = ui.Color{
						R: 32,
						G: 200,
						B: 32,
						A: 255,
					}
				}
				s.textEmail.SetValue(str)
				return true
			},
		},
	})
	s.textEmail = ui.NewTextElement(ui.TextElementConfig{
		Style: `
			Origin CenterX
			ContentOrigin CenterX CenterY
			X 50%
			Y 200
			H 20
			W 100%
		`,
	})

	elConfirm = ui.NewButtonElement(ui.ButtonElementConfig{
		Style: `
			Origin CenterX
			X 50%
			Y 230
			W 100%
			MinW 100
			MaxW 200
		`,
		Value: "Confirm",
		Events: ui.Events{
			OnMouseButtonUp: func(button uint8, x, y int32) bool {
				if !s.verifyAll() {
					s.OutputText.SetValue("Fix errors in registration form.")
				} else {
					s.OutputText.SetValue("Registering...")
				}
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
		Value: " ",
	})

	s.InputWindow.AdoptChild(s.elUsername)
	s.InputWindow.AdoptChild(s.textUsername)
	s.InputWindow.AdoptChild(s.elPassword)
	s.InputWindow.AdoptChild(s.elPassword2)
	s.InputWindow.AdoptChild(s.textPassword)
	s.InputWindow.AdoptChild(s.elEmail)
	s.InputWindow.AdoptChild(s.textEmail)
	s.InputWindow.AdoptChild(elConfirm)
	s.RegisterWindow.AdoptChild(elBack)
	s.RegisterWindow.AdoptChild(s.OutputText)

	go s.Loop()

	return
}

// Close our Register state.
func (s *Register) Close() {
	s.InputWindow.Destroy()
	s.RegisterWindow.Destroy()
}

func (s *Register) verifyUsername() (string, bool) {
	username := s.elUsername.GetValue()

	if username == "" {
		return "Username cannot be empty.", false
	}

	return "OK", true
}

func (s *Register) verifyPassword() (string, bool) {
	password := s.elPassword.GetValue()
	password2 := s.elPassword2.GetValue()

	if password == "" {
		return "Password cannot be empty.", false
	}
	if password != password2 {
		return "Passwords must be the same.", false
	}

	return "OK", true
}

func (s *Register) verifyEmail() (string, bool) {
	email := s.elEmail.GetValue()

	if email == "" {
		return "Account will not be recoverable without an email address.", false
	} else if matched, _ := regexp.Match(`^.+@([^.]*\.).*[^.]$`, []byte(email)); !matched {
		return "Email address invalid.", false
	}

	return "OK", true
}

func (s *Register) verifyAll() bool {
	if _, ok := s.verifyUsername(); !ok {
		return false
	}
	if _, ok := s.verifyPassword(); !ok {
		return false
	}

	return true
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
