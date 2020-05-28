package states

import (
	"fmt"

	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-common/network"
)

// CharacterCreation is our State for connecting as, creating, or deleting a
// character.
type CharacterCreation struct {
	client.State
	SelectionContainer  ui.Container
	CharactersContainer ui.Container
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

	s.CharactersContainer.Setup(ui.ContainerConfig{
		Style: `
			W 30%
			H 100%
			BackgroundColor 128 128 128 128
		`,
	})

	var elName, elCreate ui.ElementI

	elName = ui.NewInputElement(ui.InputElementConfig{
		Style: `
			Origin CenterX CenterY
			X 50%
			Y 10%
			H 20%
			W 100%
			MaxW 200
			MaxH 30
			MinH 25
			ForegroundColor 255 0 0 255
		`,
		Placeholder: "character name",
		Events: ui.Events{
			OnKeyDown: func(char uint8, modifiers uint16) bool {
				if char == 13 { // Enter
					elCreate.OnMouseButtonUp(1, 0, 0)
				}
				return true
			},
		},
	})

	elCreate = ui.NewButtonElement(ui.ButtonElementConfig{
		Style: `
			Origin CenterX CenterY
			X 50%
			Y 30%
			W 100%
			MaxW 200
		`,
		Value: "Create Character",
		Events: ui.Events{
			OnMouseButtonUp: func(button uint8, x int32, y int32) bool {
				s.Client.Send(network.Command(network.CommandCharacter{
					Type:       network.CreateCharacter,
					Characters: []string{elName.GetValue()},
				}))
				return false
			},
		},
	})

	// TODO:
	/*
		s.Client.Log.Printf("Logging in with dummy character")
		s.Client.Send(network.Command(network.CommandCharacter{
			Type:       network.ChooseCharacter,
			Characters: []string{"dummy"},
		}))

	*/

	s.SelectionContainer.AdoptChannel <- elName
	s.SelectionContainer.AdoptChannel <- elCreate
	s.SelectionContainer.AdoptChannel <- s.CharactersContainer.This
	s.Client.RootWindow.AdoptChannel <- s.SelectionContainer.This

	// Let the server know we're ready!
	s.Client.Send(network.Command(network.CommandBasic{
		Type: network.Okay,
	}))

	go s.Loop()

	//next = Client.StateI(&Game{})
	return
}

// addCharacter adds a button for the provided character name.
func (s *CharacterCreation) addCharacter(offset int, name string) {
	children := s.CharactersContainer.GetChildren()

	for _, child := range children {
		if _, ok := child.(*ui.ButtonElement); ok {
			offset++
		}
	}

	elChar := ui.NewButtonElement(ui.ButtonElementConfig{
		Style: fmt.Sprintf(`
			Origin CenterX CenterY
			X 50%%
			Y %d%%
			W 100%%
			MaxW 75%%
		`, 10+offset*10),
		Value: name,
		Events: ui.Events{
			OnMouseButtonUp: func(button uint8, x int32, y int32) bool {
				s.Client.Log.Printf("Logging in with character %s", name)
				s.Client.Send(network.Command(network.CommandCharacter{
					Type:       network.ChooseCharacter,
					Characters: []string{name},
				}))
				return false
			},
		},
	})
	s.CharactersContainer.AdoptChannel <- elChar
}

// Close our CharacterCreation State.
func (s *CharacterCreation) Close() {
	s.SelectionContainer.DestroyChannel <- true
}

// Loop is our loop for managing network activity and beyond.
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
		// CreateCharacter is how the server notifies us of new characters
		if t.Type == network.CreateCharacter {
			// Add character buttons.
			for i, name := range t.Characters {
				s.addCharacter(i, name)
			}
		} else if t.Type == network.ChooseCharacter {
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
