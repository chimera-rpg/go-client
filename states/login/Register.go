package login

import (
	"fmt"
	"image/color"
	"regexp"

	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-common/network"
)

// Register is a state for user Registration.
// It follows after Login and returns to Login.
type Register struct {
	client.State
	RegisterContainer, InputContainer            ui.Container
	OutputText                                   ui.ElementI
	elUsername, elPassword, elPassword2, elEmail ui.ElementI
	textUsername, textPassword, textEmail        ui.ElementI
}

// Init our Register state.
func (s *Register) Init(v interface{}) (next client.StateI, nextArgs interface{}, err error) {
	var elBack, elConfirm ui.ElementI

	err = s.RegisterContainer.Setup(ui.ContainerConfig{
		Value: "Register",
		Style: s.Client.DataManager.Styles["Register"]["Container"],
	})

	err = s.InputContainer.Setup(ui.ContainerConfig{
		Style: s.Client.DataManager.Styles["Register"]["InputsContainer"],
	})

	s.elUsername = ui.NewInputElement(ui.InputElementConfig{
		Style:       s.Client.DataManager.Styles["Register"]["UsernameInput"],
		Placeholder: "username",
		Events: ui.Events{
			OnAdopted: func(parent ui.ElementI) {
				s.elUsername.Focus()
			},
			OnChange: func() bool {
				str, ok := s.verifyUsername()
				if !ok {
					s.textUsername.GetStyle().ForegroundColor = color.NRGBA{
						R: 255,
						G: 32,
						B: 32,
						A: 255,
					}
				} else {
					s.textUsername.GetStyle().ForegroundColor = color.NRGBA{
						R: 32,
						G: 255,
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
		Style: s.Client.DataManager.Styles["Register"]["UsernameResult"],
	})

	s.elPassword = ui.NewInputElement(ui.InputElementConfig{
		Style:       s.Client.DataManager.Styles["Register"]["PasswordInput"],
		Password:    true,
		Placeholder: "password",
		Events: ui.Events{
			OnChange: func() bool {
				str, ok := s.verifyPassword()
				if !ok {
					s.textPassword.GetStyle().ForegroundColor = color.NRGBA{
						R: 255,
						G: 32,
						B: 32,
						A: 255,
					}
				} else {
					s.textPassword.GetStyle().ForegroundColor = color.NRGBA{
						R: 32,
						G: 255,
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
		Style:       s.Client.DataManager.Styles["Register"]["PasswordConfirmInput"],
		Password:    true,
		Placeholder: "confirm password",
		Events: ui.Events{
			OnChange: func() bool {
				str, ok := s.verifyPassword()
				if !ok {
					s.textPassword.GetStyle().ForegroundColor = color.NRGBA{
						R: 255,
						G: 32,
						B: 32,
						A: 255,
					}
				} else {
					s.textPassword.GetStyle().ForegroundColor = color.NRGBA{
						R: 32,
						G: 255,
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
		Style: s.Client.DataManager.Styles["Register"]["PasswordResult"],
	})

	s.elEmail = ui.NewInputElement(ui.InputElementConfig{
		Style:       s.Client.DataManager.Styles["Register"]["EmailInput"],
		Placeholder: "email",
		Events: ui.Events{
			OnChange: func() bool {
				str, ok := s.verifyEmail()
				if !ok {
					s.textEmail.GetStyle().ForegroundColor = color.NRGBA{
						R: 200,
						G: 200,
						B: 32,
						A: 255,
					}
				} else {
					s.textEmail.GetStyle().ForegroundColor = color.NRGBA{
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
		Style: s.Client.DataManager.Styles["Register"]["EmailResult"],
	})

	elConfirm = ui.NewButtonElement(ui.ButtonElementConfig{
		Style: s.Client.DataManager.Styles["Register"]["ConfirmButton"],
		Value: "Confirm",
		Events: ui.Events{
			OnPressed: func(button uint8, x, y int32) bool {
				if !s.verifyAll() {
					s.OutputText.SetValue("Fix errors in registration form.")
				} else {
					s.OutputText.SetValue("Registering...")
					s.Client.Send(network.Command(network.CommandLogin{
						Type:  network.Register,
						User:  s.elUsername.GetValue(),
						Pass:  s.elPassword.GetValue(),
						Email: s.elEmail.GetValue(),
					}))
				}
				return true
			},
		},
	})

	elBack = ui.NewButtonElement(ui.ButtonElementConfig{
		Style: s.Client.DataManager.Styles["Register"]["BackButton"],
		Value: "Back",
		Events: ui.Events{
			OnPressed: func(button uint8, x, y int32) bool {
				s.Client.StateChannel <- client.StateMessage{Pop: true, Args: nil}
				return true
			},
		},
	})

	s.OutputText = ui.NewTextElement(ui.TextElementConfig{
		Style: s.Client.DataManager.Styles["Register"]["OutputText"],
		Value: " ",
	})

	s.InputContainer.AdoptChannel <- s.elUsername
	s.InputContainer.AdoptChannel <- s.textUsername
	s.InputContainer.AdoptChannel <- s.elPassword
	s.InputContainer.AdoptChannel <- s.elPassword2
	s.InputContainer.AdoptChannel <- s.textPassword
	s.InputContainer.AdoptChannel <- s.elEmail
	s.InputContainer.AdoptChannel <- s.textEmail

	s.RegisterContainer.AdoptChannel <- s.InputContainer.This
	s.RegisterContainer.AdoptChannel <- elBack
	s.RegisterContainer.AdoptChannel <- elConfirm
	s.RegisterContainer.AdoptChannel <- s.OutputText

	s.Client.RootWindow.AdoptChannel <- s.RegisterContainer.This

	go s.Loop()

	return
}

// Close our Register state.
func (s *Register) Close() {
	s.InputContainer.DestroyChannel <- true
	s.RegisterContainer.DestroyChannel <- true
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
		case <-s.CloseChan:
			return
		case cmd := <-s.Client.CmdChan:
			s.HandleNet(cmd)
		case <-s.Client.ClosedChan:
			s.Client.StateChannel <- client.StateMessage{PopToTop: true, Args: nil}
			return
		}
	}
}

// HandleNet handles the network commands received in Loop().
func (s *Register) HandleNet(cmd network.Command) bool {
	switch t := cmd.(type) {
	case network.CommandBasic:
		if t.Type == network.Reject {
			s.OutputText.GetUpdateChannel() <- ui.UpdateValue{Value: t.String}
		} else if t.Type == network.Okay {
			s.Client.StateChannel <- client.StateMessage{State: &Login{}, Args: LoginState{defaultState, s.elUsername.GetValue(), s.elPassword.GetValue(), t.String}}
			return true
		}
	default:
		msg := fmt.Sprintf("Server sent non CommandBasic")
		s.Client.Log.Print(msg)
		s.Client.StateChannel <- client.StateMessage{PopToTop: true, Args: msg}
		return true
	}
	return false
}
