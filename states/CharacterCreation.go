package states

import (
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-common/network"
)

// CharacterCreation is our State for connecting as, creating, or deleting a
// character.
type CharacterCreation struct {
	client.State
	SelectionWindow ui.Window
	CharacterWindow ui.Window
}

// Init is our CharacterCreation init state.
func (s *CharacterCreation) Init(t interface{}) (next client.StateI, nextArgs interface{}, err error) {
	s.Client.Log.Print("CharacterCreation State")

	err = s.SelectionWindow.Setup(ui.WindowConfig{
		Value: "Selection",
		Style: `
			X 5%
			Y 5%
			W 90%%
			H 20%
			Origin CenterX CenterY
		`,
		Parent: s.Client.RootWindow,
		RenderFunc: func(w *ui.Window) {
			w.Context.Renderer.SetDrawColor(32, 32, 128, 128)
			w.Context.Renderer.Clear()
		},
	})

	elSelection := ui.NewTextElement(ui.TextElementConfig{
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
		Events: ui.Events{
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

	s.SelectionWindow.AdoptChild(elSelection)

	go s.Loop()
	/*for {
		cmd := <-s.Client.CmdChan
		switch t := cmd.(type) {
		case network.CommandBasic:
			if t.Type == network.REJECT {
				s.Client.Log.Printf("Server rejected us: %s\n", t.String)
			} else if t.Type == network.OK {
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

// Close our CharacterCreation State.
func (s *CharacterCreation) Close() {
	s.SelectionWindow.Destroy()
}

// Loop is our loop for managing network activitiy and beyond.
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
			s.Client.StateChannel <- client.StateMessage{State: &List{}, Args: nil}
			return
		}
	}
}

// HandleNet manages our network communications.
func (s *CharacterCreation) HandleNet(cmd network.Command) bool {
	switch t := cmd.(type) {
	case network.CommandBasic:
		if t.Type == network.REJECT {
			s.Client.Log.Printf("Server rejected us: %s\n", t.String)
		} else if t.Type == network.OK {
			s.Client.Log.Printf("Server accepted us: %s\n", t.String)
			s.Client.StateChannel <- client.StateMessage{State: &Game{}, Args: nil}
			return true
		}
	default:
		s.Client.Log.Print("Server sent non CommandBasic")
		s.Client.StateChannel <- client.StateMessage{State: &List{}, Args: nil}
		return true
	}
	return false
}
