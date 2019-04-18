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

func (s *Login) Init(t interface{}) (next Client.StateI, nextArgs interface{}, err error) {
	err = s.LoginWindow.Setup(UI.WindowConfig{
		Value: "Selection",
		Style: UI.Style{
			X: UI.Number{
				Value:      50,
				Percentage: true,
			},
			Y: UI.Number{
				Value: 120,
			},
			W: UI.Number{
				Percentage: true,
				Value:      70,
			},
			H: UI.Number{
				Percentage: true,
				Value:      20,
			},
			Origin: UI.ORIGIN_CENTERX | UI.ORIGIN_CENTERY,
		},
		Parent: s.Client.RootWindow,
		RenderFunc: func(w *UI.Window) {
			w.Context.Renderer.SetDrawColor(32, 32, 128, 128)
			w.Context.Renderer.Clear()
		},
	})

	el_username := UI.NewInputElement(UI.InputElementConfig{
		Style: UI.Style{
			ForegroundColor: UI.Color{255, 255, 255, 255, true},
			BackgroundColor: UI.Color{0, 0, 0, 128, true},
			PaddingLeft: UI.Number{
				Percentage: true,
				Value:      5,
			},
			PaddingRight: UI.Number{
				Percentage: true,
				Value:      5,
			},
			PaddingTop: UI.Number{
				Percentage: true,
				Value:      5,
			},
			PaddingBottom: UI.Number{
				Percentage: true,
				Value:      5,
			},
			Origin: UI.ORIGIN_CENTERX | UI.ORIGIN_CENTERY,
			X: UI.Number{
				Value:      50,
				Percentage: true,
			},
			Y: UI.Number{
				Value:      30,
				Percentage: true,
			},
			H: UI.Number{
				Value:      20,
				Percentage: true,
			},
		},
		Value: "username",
		Events: UI.Events{
			OnMouseMove: func(x int32, y int32) bool {
				s.Client.Log.Printf("Movement: %dx%d! :)\n", x, y)
				return false
			},
			OnMouseButtonDown: func(button uint8, x int32, y int32) bool {
				s.Client.Log.Printf("Clicky use: %d @ %dx%d! :D\n", button, x, y)
				return false
			},
			OnMouseIn: func(x int32, y int32) bool {
				s.Client.Log.Printf("MouseIn\n")
				return false
			},
			OnMouseOut: func(x int32, y int32) bool {
				s.Client.Log.Printf("MouseOut\n")
				return false
			},
		},
	})
	el_password := UI.NewInputElement(UI.InputElementConfig{
		Style: UI.Style{
			ForegroundColor: UI.Color{255, 255, 255, 255, true},
			BackgroundColor: UI.Color{0, 0, 0, 128, true},
			//Position: UI.POSITION_RELATIVE,
			PaddingLeft: UI.Number{
				Percentage: true,
				Value:      5,
			},
			PaddingRight: UI.Number{
				Percentage: true,
				Value:      5,
			},
			PaddingTop: UI.Number{
				Percentage: true,
				Value:      5,
			},
			PaddingBottom: UI.Number{
				Percentage: true,
				Value:      5,
			},
			Origin: UI.ORIGIN_CENTERX | UI.ORIGIN_CENTERY,
			X: UI.Number{
				Value:      50,
				Percentage: true,
			},
			Y: UI.Number{
				Value:      70,
				Percentage: true,
			},
			H: UI.Number{
				Value:      20,
				Percentage: true,
			},
		},
		Value: "password",
		Events: UI.Events{
			OnMouseMove: func(x int32, y int32) bool {
				s.Client.Log.Printf("Movement: %dx%d! :)\n", x, y)
				return false
			},
			OnMouseButtonDown: func(button uint8, x int32, y int32) bool {
				s.Client.Log.Printf("Clicky: %d @ %dx%d! :D\n", button, x, y)
				return false
			},
			OnMouseIn: func(x int32, y int32) bool {
				s.Client.Log.Printf("MouseIn\n")
				return false
			},
			OnMouseOut: func(x int32, y int32) bool {
				s.Client.Log.Printf("MouseOut\n")
				return false
			},
		},
	})

	s.LoginWindow.AdoptChild(el_username)
	s.LoginWindow.AdoptChild(el_password)

	s.Client.Log.Print("Login State")
	// Show UI for Username/Password input:
	//   * Main: Server Info Panel, Username, Password, Login, Register
	//     * Register: Enter Password Again, E-Mail(optional field)
	//       * User exists! (go back to Main)
	//       * Registered! (go back to Main w/ Login prefilled)
	//     * Login
	//       * Success! (go to Character Selection/Creation State)
	//       * Bad password/username! (go back to Main)
	/*s.Client.Send(Net.Command(Net.CommandLogin{
		Type: Net.LOGIN,
		User: "nommak",
		Pass: "nommak",
	}))*/

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
				return
			}
		case <-s.Client.ClosedChan:
			s.Client.Log.Print("Lost connection to server.")
			s.Client.StateChannel <- Client.StateMessage{&List{}, nil}
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
