package login

import (
	"flag"
	"fmt"

	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/states/game"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-server/network"
)

// CharacterSelection is our State for connecting as, creating, or deleting a
// character.
type CharacterSelection struct {
	client.State
	layout ui.LayoutEntry
	bail   chan bool
}

// Init is our CharacterSelection init state.
func (s *CharacterSelection) Init(t interface{}) (next client.StateI, nextArgs interface{}, err error) {
	s.bail = make(chan bool)
	s.Client.Log.Print("CharacterSelection State")

	s.layout = s.Client.DataManager.Layouts["Selection"][0].Generate(s.Client.DataManager.Styles["Selection"], map[string]interface{}{
		"Container": ui.ContainerConfig{
			Value: "Selection",
		},
		"Characters": ui.ContainerConfig{
			Value: "Character",
		},
		"CharacterName": ui.InputElementConfig{
			Placeholder: "character name",
			Events: ui.Events{
				OnKeyDown: func(char uint8, modifiers uint16, repeat bool) bool {
					if char == 13 { // Enter
						s.layout.Find("CreateButton").Element.OnPressed(1, 0, 0)
					}
					return true
				},
			},
		},
		"CreateButton": ui.ButtonElementConfig{
			Value: "Create Character",
			Events: ui.Events{
				OnPressed: func(button uint8, x int32, y int32) bool {
					s.bail <- true
					s.Client.StateChannel <- client.StateMessage{Push: true, State: &CharacterCreation{}}
					return false
				},
			},
		},
	})

	s.Client.RootWindow.AdoptChannel <- s.layout.Find("Container").Element

	// Let the server know we're ready!
	s.Client.Send(network.Command(network.CommandCharacter{
		Type: network.QueryCharacters,
	}))

	go s.Loop()

	return
}

// addCharacter adds a button for the provided character name.
func (s *CharacterSelection) addCharacter(offset int, name string) {
	children := s.layout.Find("Characters").Element.GetChildren()

	for _, child := range children {
		if _, ok := child.(*ui.ButtonElement); ok {
			offset++
		}
	}

	isFocused := false
	if name == s.Client.DataManager.Config.Servers[s.Client.CurrentServer].Character {
		isFocused = true
	}

	elChar := ui.NewButtonElement(ui.ButtonElementConfig{
		Style: fmt.Sprintf(s.Client.DataManager.Styles["Selection"]["CharacterEntry_fmt"], 10+offset*10),
		Value: name,
		Events: ui.Events{
			OnPressed: func(button uint8, x int32, y int32) bool {
				s.Client.Log.Printf("Logging in with character %s", name)
				s.Client.DataManager.Config.Servers[s.Client.CurrentServer].Character = name
				s.Client.Send(network.Command(network.CommandCharacter{
					Type:       network.ChooseCharacter,
					Characters: []string{name},
				}))
				return false
			},
		},
	})
	if isFocused {
		elChar.Focus()
	}
	s.layout.Find("Characters").Element.GetAdoptChannel() <- elChar
}

// Close our CharacterSelection State.
func (s *CharacterSelection) Close() {
	s.layout.Find("Container").Element.GetDestroyChannel() <- true
}

func (s *CharacterSelection) Leave() {
	s.layout.Find("Container").Element.GetUpdateChannel() <- ui.UpdateHidden(true)
}

func (s *CharacterSelection) Enter(args ...interface{}) {
	s.layout.Find("Container").Element.GetUpdateChannel() <- ui.UpdateHidden(false)
	go s.Loop()
}

// Loop is our loop for managing network activity and beyond.
func (s *CharacterSelection) Loop() {
	// Attempt to use provided character.
	character := flag.Lookup("character")
	if character.Value.String() != character.DefValue {
		s.Client.Send(network.Command(network.CommandCharacter{
			Type:       network.ChooseCharacter,
			Characters: []string{character.Value.String()},
		}))
	}

	for {
		if !s.Running {
			continue
		}
		select {
		case <-s.bail:
			return
		case <-s.Client.DataManager.UpdatedImageIDs:
			// TODO: Refresh character image
		case cmd := <-s.Client.CmdChan:
			ret := s.HandleNet(cmd)
			if ret {
				return
			}
		case <-s.Client.ClosedChan:
			s.Client.Log.Print("Lost connection to server.")
			s.Client.StateChannel <- client.StateMessage{PopToTop: true, Args: nil}
			return
		}
	}
}

// HandleNet manages our network communications.
func (s *CharacterSelection) HandleNet(cmd network.Command) bool {
	switch t := cmd.(type) {
	case network.CommandGraphics:
		s.Client.DataManager.HandleGraphicsCommand(t)
	case network.CommandAnimation:
		s.Client.DataManager.HandleAnimationCommand(t)
	case network.CommandBasic:
		if t.Type == network.Reject {
			s.Client.Log.Printf("Server rejected us: %s\n", t.String)
		} else if t.Type == network.Okay {
			s.Client.Log.Printf("Server accepted us: %s\n", t.String)
			// Might as well save the configuration now.
			if err := s.Client.DataManager.Config.Write(); err != nil {
				s.Client.Log.Errorln(err)
			}
			s.Client.StateChannel <- client.StateMessage{Push: true, State: &game.Game{}, Args: nil}
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
			// ChooseCharacter is how the server lets us know we're logging in as a character.
			s.Client.StateChannel <- client.StateMessage{Push: true, State: &game.Game{}, Args: nil}
			// Might as well save the configuration now.
			if err := s.Client.DataManager.Config.Write(); err != nil {
				s.Client.Log.Errorln(err)
			}
			return true
		} else {
			s.Client.Log.Printf("Unhandled CommandCharacter type %d\n", t.Type)
		}
	default:
		s.Client.Log.Printf("Server sent non CommandBasic\n")
		s.Client.StateChannel <- client.StateMessage{PopToTop: true, Args: nil}
		return true
	}
	return false
}
