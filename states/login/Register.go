package login

import (
	"fmt"
	"image/color"
	"regexp"

	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-server/network"
)

// Register is a state for user Registration.
// It follows after Login and returns to Login.
type Register struct {
	client.State
	layout ui.LayoutEntry
}

// Init our Register state.
func (s *Register) Init(v interface{}) (next client.StateI, nextArgs interface{}, err error) {

	s.layout = s.Client.DataManager.Layouts["Register"][0].Generate(s.Client.DataManager.Styles["Register"], map[string]interface{}{
		"Container": ui.ContainerConfig{
			Value: "Register",
			Style: s.Client.DataManager.Styles["Register"]["Container"],
		},
		//"InputsContainer": ui.ContainerConfig{}
		"UsernameInput": ui.InputElementConfig{
			Placeholder: "username",
			Events: ui.Events{
				OnAdopted: func(parent ui.ElementI) {
					s.layout.Find("UsernameInput").Element.Focus()
				},
				OnChange: func() bool {
					textUsername := s.layout.Find("UsernameResult").Element
					str, ok := s.verifyUsername()
					if !ok {
						textUsername.GetStyle().ForegroundColor = color.NRGBA{
							R: 255,
							G: 32,
							B: 32,
							A: 255,
						}
					} else {
						textUsername.GetStyle().ForegroundColor = color.NRGBA{
							R: 32,
							G: 255,
							B: 32,
							A: 255,
						}
					}
					textUsername.SetValue(str)
					return true
				},
			},
		},
		//"UsernameResult": ui.TextElementConfig{}
		"PasswordInput": ui.InputElementConfig{
			Password:    true,
			Placeholder: "password",
			Events: ui.Events{
				OnChange: func() bool {
					str, ok := s.verifyPassword()
					textPassword := s.layout.Find("PasswordResult").Element
					if !ok {
						textPassword.GetStyle().ForegroundColor = color.NRGBA{
							R: 255,
							G: 32,
							B: 32,
							A: 255,
						}
					} else {
						textPassword.GetStyle().ForegroundColor = color.NRGBA{
							R: 32,
							G: 255,
							B: 32,
							A: 255,
						}
					}
					textPassword.SetValue(str)
					return true
				},
			},
		},
		"PasswordConfirmInput": ui.InputElementConfig{
			Password:    true,
			Placeholder: "confirm password",
			Events: ui.Events{
				OnChange: func() bool {
					str, ok := s.verifyPassword()
					textPassword := s.layout.Find("PasswordResult").Element
					if !ok {
						textPassword.GetStyle().ForegroundColor = color.NRGBA{
							R: 255,
							G: 32,
							B: 32,
							A: 255,
						}
					} else {
						textPassword.GetStyle().ForegroundColor = color.NRGBA{
							R: 32,
							G: 255,
							B: 32,
							A: 255,
						}
					}
					textPassword.SetValue(str)
					return true
				},
			},
		},
		//"PasswordResult": ui.TextElementConfig{}
		"EmailInput": ui.InputElementConfig{
			Placeholder: "email",
			Events: ui.Events{
				OnChange: func() bool {
					str, ok := s.verifyEmail()
					textEmail := s.layout.Find("EmailResult").Element
					if !ok {
						textEmail.GetStyle().ForegroundColor = color.NRGBA{
							R: 200,
							G: 200,
							B: 32,
							A: 255,
						}
					} else {
						textEmail.GetStyle().ForegroundColor = color.NRGBA{
							R: 32,
							G: 200,
							B: 32,
							A: 255,
						}
					}
					textEmail.SetValue(str)
					return true
				},
			},
		},
		// "EmailResult"
		"ConfirmButton": ui.ButtonElementConfig{
			Value: "Confirm",
			Events: ui.Events{
				OnPressed: func(button uint8, x, y int32) bool {
					outputText := s.layout.Find("OutputText").Element
					if !s.verifyAll() {
						outputText.SetValue("Fix errors in registration form.")
					} else {
						outputText.SetValue("Registering...")
						s.Client.Send(network.Command(network.CommandLogin{
							Type:  network.Register,
							User:  s.layout.Find("UsernameInput").Element.GetValue(),
							Pass:  s.layout.Find("PasswordInput").Element.GetValue(),
							Email: s.layout.Find("EmailInput").Element.GetValue(),
						}))
					}
					return true
				},
			},
		},
		"BackButton": ui.ButtonElementConfig{
			Value: "Back",
			Events: ui.Events{
				OnPressed: func(button uint8, x, y int32) bool {
					s.Client.StateChannel <- client.StateMessage{Pop: true, Args: nil}
					return true
				},
			},
		},
		"OutputText": ui.TextElementConfig{
			Value: " ",
		},
	})

	s.Client.RootWindow.AdoptChannel <- s.layout.Find("Container").Element

	go s.Loop()

	return
}

// Close our Register state.
func (s *Register) Close() {
	s.layout.Find("Container").Element.GetDestroyChannel() <- true
}

func (s *Register) verifyUsername() (string, bool) {
	username := s.layout.Find("UsernameInput").Element.GetValue()

	if username == "" {
		return "Username cannot be empty.", false
	}

	return "OK", true
}

func (s *Register) verifyPassword() (string, bool) {
	password := s.layout.Find("PasswordInput").Element.GetValue()
	password2 := s.layout.Find("PasswordConfirmInput").Element.GetValue()

	if password == "" {
		return "Password cannot be empty.", false
	}
	if password != password2 {
		return "Passwords must be the same.", false
	}

	return "OK", true
}

func (s *Register) verifyEmail() (string, bool) {
	email := s.layout.Find("EmailInput").Element.GetValue()

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
			s.layout.Find("OutputText").Element.GetUpdateChannel() <- ui.UpdateValue{Value: t.String}
		} else if t.Type == network.Okay {
			s.Client.StateChannel <- client.StateMessage{State: &Login{}, Args: LoginState{defaultState, s.layout.Find("UsernameInput").Element.GetValue(), s.layout.Find("PasswordInput").Element.GetValue(), t.String}}
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
