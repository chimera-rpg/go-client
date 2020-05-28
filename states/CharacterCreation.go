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
	SelectionContainer ui.Container
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

	var elName, elSelection, elCreate ui.ElementI

	elSelection = ui.NewTextElement(ui.TextElementConfig{
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
		Password:    true,
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

	s.SelectionContainer.AdoptChannel <- elSelection
	s.SelectionContainer.AdoptChannel <- elName
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
		// CreateCharacter is how the server notifies us of new characters
		if t.Type == network.CreateCharacter {
			// FIXME: This is temporary so we can actually login with characters.
			for i, name := range t.Characters {
				go func(i int, name string) {
					elChar := ui.NewButtonElement(ui.ButtonElementConfig{
						Style: fmt.Sprintf(`
							Origin CenterX CenterY
							X 25%%
							Y %d%%
							W 100%%
							MaxW 25%%
						`, 10+i*10),
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

					s.SelectionContainer.AdoptChannel <- elChar
				}(i, name)
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
