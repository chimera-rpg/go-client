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
	SelectionContainer ui.Container
	CharacterContainer ui.Container
}

// Init is our CharacterCreation init state.
func (s *CharacterCreation) Init(t interface{}) (next client.StateI, nextArgs interface{}, err error) {
	s.Client.Log.Print("CharacterCreation State")

	err = s.SelectionContainer.Setup(ui.ContainerConfig{
		Value: "Selection",
		Style: `
			W 100%
			H 100%
		`,
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

	elCreate := ui.NewButtonElement(ui.ButtonElementConfig{
		Style: `
			Origin CenterX CenterY
			X 50%
			Y 30%
			W 100%
			MaxW 200
		`,
		Value: "Dummy Character",
		Events: ui.Events{
			OnMouseButtonDown: func(button uint8, x int32, y int32) bool {
				s.Client.Log.Printf("Logging in with dummy character")
				s.Client.Send(network.Command(network.CommandCharacter{
					Type:       network.ChooseCharacter,
					Characters: []string{"dummy"},
				}))
				return false
			},
		},
	})

	s.SelectionContainer.AdoptChannel <- elSelection
	s.SelectionContainer.AdoptChannel <- elCreate
	s.Client.RootWindow.AdoptChannel <- s.SelectionContainer.This

	go s.Loop()

	//next = Client.StateI(&Game{})
	return
}

// Close our CharacterCreation State.
func (s *CharacterCreation) Close() {
	s.SelectionContainer.DestroyChannel <- true
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
		if t.Type == network.Reject {
			s.Client.Log.Printf("Server rejected us: %s\n", t.String)
		} else if t.Type == network.Okay {
			s.Client.Log.Printf("Server accepted us: %s\n", t.String)
			s.Client.StateChannel <- client.StateMessage{State: &Game{}, Args: nil}
			return true
		}
	case network.CommandCharacter:
		if t.Type == network.ChooseCharacter {
			s.Client.StateChannel <- client.StateMessage{State: &Game{}, Args: nil}
			return true
		} else {
			s.Client.Log.Printf("Unhandled CommandCharacter type %d\n", t.Type)
		}
	default:
		s.Client.Log.Printf("Server sent non CommandBasic\n")
		s.Client.StateChannel <- client.StateMessage{State: &List{}, Args: nil}
		return true
	}
	return false
}
