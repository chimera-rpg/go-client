package States

import (
	"github.com/chimera-rpg/go-client/Client"
	"github.com/chimera-rpg/go-client/UI"
	"github.com/chimera-rpg/go-common/Net"
)

type CharacterCreation struct {
	Client.State
	SelectionWindow UI.Window
	CharacterWindow UI.Window
}

func (s *CharacterCreation) Init(t interface{}) (next Client.StateI, nextArgs interface{}, err error) {
	s.Client.Log.Print("CharacterCreation State")

	err = s.SelectionWindow.Setup(UI.WindowConfig{
		Value: "Selection",
		Style: UI.Style{
			X: UI.Number{
				Value: 8,
			},
			Y: UI.Number{
				Value: 8,
			},
			W: UI.Number{
				Percentage: true,
				Value:      70,
			},
			H: UI.Number{
				Percentage: true,
				Value:      20,
			},
		},
		Parent: s.Client.RootWindow,
		RenderFunc: func(w *UI.Window) {
			w.Context.Renderer.SetDrawColor(32, 32, 128, 128)
			w.Context.Renderer.Clear()
		},
	})

	el_selection := UI.NewInputElement(UI.InputElementConfig{
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
				Value:      10,
				Percentage: true,
			},
		},
		Value: "Select your Character",
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

	s.SelectionWindow.AdoptChild(el_selection)

	go s.Loop()
	/*for {
		cmd := <-s.Client.CmdChan
		switch t := cmd.(type) {
		case Net.CommandBasic:
			if t.Type == Net.REJECT {
				s.Client.Log.Printf("Server rejected us: %s\n", t.String)
			} else if t.Type == Net.OK {
				s.Client.Log.Printf("Server accepted us: %s\n", t.String)
				break
			}
		default:
			s.Client.Log.Print("Server sent non CommandBasic")
			next = Client.StateI(&List{})
			return
		}
	}*/

	//next = Client.StateI(&Game{})
	return
}

func (s *CharacterCreation) Close() {
	s.SelectionWindow.Destroy()
}

func (s *CharacterCreation) Loop() {
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

func (s *CharacterCreation) HandleNet(cmd Net.Command) bool {
	switch t := cmd.(type) {
	case Net.CommandBasic:
		if t.Type == Net.REJECT {
			s.Client.Log.Printf("Server rejected us: %s\n", t.String)
		} else if t.Type == Net.OK {
			s.Client.Log.Printf("Server accepted us: %s\n", t.String)
			s.Client.StateChannel <- Client.StateMessage{&Game{}, nil}
			return true
		}
	default:
		s.Client.Log.Print("Server sent non CommandBasic")
		s.Client.StateChannel <- Client.StateMessage{&List{}, nil}
		return true
	}
	return false
}
