package login

import (
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/states/game"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-server/network"
)

// CharacterCreation is our State for connecting as, creating, or deleting a
// character.
type CharacterCreation struct {
	client.State
	layout ui.LayoutEntry
	bail   chan bool
}

// Init is our CharacterCreation init state.
func (s *CharacterCreation) Init(t interface{}) (next client.StateI, nextArgs interface{}, err error) {
	s.bail = make(chan bool)
	s.Client.Log.Print("CharacterCreation State")

	s.layout = s.Client.DataManager.Layouts["Creation"][0].Generate(s.Client.DataManager.Styles["Creation"], map[string]interface{}{
		"Container": ui.ContainerConfig{
			Value: "Creation",
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
					s.Client.Send(network.Command(network.CommandCreateCharacter{
						Name: s.layout.Find("CharacterName").Element.GetValue(),
					}))
					return false
				},
			},
		},
		"BackButton": ui.ButtonElementConfig{
			Value: "Back",
			Events: ui.Events{
				OnPressed: func(button uint8, x int32, y int32) bool {
					s.bail <- true
					s.Client.StateChannel <- client.StateMessage{Pop: true}
					return false
				},
			},
		},
	})

	s.Client.RootWindow.AdoptChannel <- s.layout.Find("Container").Element

	// Let the server know we're ready!
	s.Client.Send(network.Command(network.CommandQueryGenera{}))

	go s.Loop()

	return
}

// addCharacter adds a button for the provided character name.
func (s *CharacterCreation) addCharacter(name string) {
	isFocused := false
	if name == s.Client.DataManager.Config.Servers[s.Client.CurrentServer].Character {
		isFocused = true
	}

	elChar := ui.NewButtonElement(ui.ButtonElementConfig{
		Style: s.Client.DataManager.Styles["Selection"]["CharacterEntry"],
		Value: name,
		Events: ui.Events{
			OnPressed: func(button uint8, x int32, y int32) bool {
				s.Client.Log.Printf("Logging in with character %s", name)
				s.Client.DataManager.Config.Servers[s.Client.CurrentServer].Character = name
				s.Client.Send(network.Command(network.CommandSelectCharacter{
					Name: name,
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

// Close our CharacterCreation State.
func (s *CharacterCreation) Close() {
	s.layout.Find("Container").Element.GetDestroyChannel() <- true
}

func (s *CharacterCreation) Leave() {
	s.layout.Find("Container").Element.GetUpdateChannel() <- ui.UpdateHidden(true)
}

func (s *CharacterCreation) Enter(args ...interface{}) {
	s.layout.Find("Container").Element.GetUpdateChannel() <- ui.UpdateHidden(false)
}

// Loop is our loop for managing network activity and beyond.
func (s *CharacterCreation) Loop() {
	for {
		if !s.Running {
			continue
		}
		select {
		case <-s.bail:
			return
		case cmd := <-s.Client.CmdChan:
			ret := s.HandleNet(cmd)
			if ret {
				return
			}
		case <-s.Client.DataManager.UpdatedImageIDs:
			// TODO: Refresh genus/species/pc image
		case <-s.Client.ClosedChan:
			s.Client.Log.Print("Lost connection to server.")
			s.Client.StateChannel <- client.StateMessage{PopToTop: true, Args: nil}
			return
		}
	}
}

// HandleNet manages our network communications.
func (s *CharacterCreation) HandleNet(cmd network.Command) bool {
	switch t := cmd.(type) {
	case network.CommandGraphics:
		s.Client.DataManager.HandleGraphicsCommand(t)
	case network.CommandAnimation:
		s.Client.DataManager.HandleAnimationCommand(t)
	case network.CommandSound:
		s.Client.DataManager.HandleSoundCommand(t)
	case network.CommandAudio:
		s.Client.DataManager.HandleAudioCommand(t)
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
	case network.CommandCreateCharacter:
		s.addCharacter(t.Name)
	case network.CommandQueryGenera:
		s.Client.Log.Println("TODO: Handle CommandGenera", t.Genera)
		for _, genus := range t.Genera {
			s.Client.DataManager.EnsureAnimation(genus.AnimationID)
		}
	case network.CommandQuerySpecies:
		s.Client.Log.Println("TODO: Handle CommandSpecies", t.Genus, t.Species)
		for _, species := range t.Species {
			s.Client.DataManager.EnsureAnimation(species.AnimationID)
		}
	default:
		s.Client.Log.Printf("Server sent incorrect Command\n")
		s.Client.StateChannel <- client.StateMessage{PopToTop: true, Args: nil}
		return true
	}
	return false
}
