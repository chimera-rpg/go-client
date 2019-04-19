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
		Style: `
			X 5%
			Y 5%
			W 90%%
			H 20%
			Origin CenterX CenterY
		`,
		Parent: s.Client.RootWindow,
		RenderFunc: func(w *UI.Window) {
			w.Context.Renderer.SetDrawColor(32, 32, 128, 128)
			w.Context.Renderer.Clear()
		},
	})

	el_selection := UI.NewTextElement(UI.TextElementConfig{
		Style: `
			PaddingLeft 5%
			PaddingRight 5%
			PaddingTop 5%
			PaddingBottom 5%
			Origin CenterX CenterY
			ForegroundColor 255 255 255 255
			BackgroundColor 0 0 0 255
			X 50%
			Y 10%
		`,
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
