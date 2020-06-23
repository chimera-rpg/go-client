package states

import (
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-client/world"
	"github.com/chimera-rpg/go-common/network"
)

// Game is our live Game state, used once the user has connected to the server
// and joined as a player character.
type Game struct {
	client.State
	GameContainer   ui.Container
	ChatWindow      ui.Container
	MapWindow       ui.Container
	InventoryWindow ui.Container
	GroundWindow    ui.Container
	StatsWindow     ui.Container
	StateWindow     ui.Container
	world           world.World
}

// Init our Game state.
func (s *Game) Init(t interface{}) (state client.StateI, nextArgs interface{}, err error) {
	// Initialize our world.
	s.world.Init(s.Client.DataManager, s.Client.Log)

	s.Client.Log.Print("Game State")

	// Main Container
	err = s.GameContainer.Setup(ui.ContainerConfig{
		Value: "Game",
		Style: `
			W 100%
			H 100%
			BackgroundColor 139 186 139 255
		`,
		Events: ui.Events{
			OnKeyDown: func(char uint8, modifiers uint16) bool {
				// TODO: Forward this to a full bind/keypress handler system.
				if char == 107 || char == 82 { // up
					s.Client.Log.Println("send north")
					s.Client.Send(network.CommandCmd{
						Cmd: network.North,
					})
				} else if char == 106 || char == 81 { // down
					s.Client.Log.Println("send south")
					s.Client.Send(network.CommandCmd{
						Cmd: network.South,
					})
				} else if char == 104 || char == 80 { // left
					s.Client.Log.Println("send west")
					s.Client.Send(network.CommandCmd{
						Cmd: network.West,
					})
				} else if char == 108 || char == 79 { // right
					s.Client.Log.Println("send east")
					s.Client.Send(network.CommandCmd{
						Cmd: network.East,
					})
				}
				return true
			},
		},
	})
	s.Client.RootWindow.AdoptChannel <- s.GameContainer.This

	// Sub-window: map
	err = s.MapWindow.Setup(ui.ContainerConfig{
		Value: "Map",
		Style: `
			X 50%
			Y 50%
			W 100%
			H 100%
			BackgroundColor 0 0 0 255
			Origin CenterX CenterY
		`,
	})
	mapText := ui.NewTextElement(ui.TextElementConfig{
		Value: "Map",
	})
	s.MapWindow.AdoptChannel <- mapText
	s.GameContainer.AdoptChannel <- s.MapWindow.This
	// Sub-window: chat
	err = s.ChatWindow.Setup(ui.ContainerConfig{
		Value: "Chat",
		Style: `
			X 8
			Y 8
			W 70%
			H 20%
			BackgroundColor 0 0 128 128
		`,
	})
	s.GameContainer.AdoptChannel <- s.ChatWindow.This
	// Sub-window: inventory
	err = s.InventoryWindow.Setup(ui.ContainerConfig{
		Value: "Inventory",
		Style: `
			X 50%
			Y 50%
			W 50%
			H 80%
			Origin CenterX CenterY
			BackgroundColor 0 128 0 128
		`,
	})
	s.GameContainer.AdoptChannel <- s.InventoryWindow.This
	s.InventoryWindow.SetHidden(true)
	// Sub-window: ground
	err = s.GroundWindow.Setup(ui.ContainerConfig{
		Value: "Ground",
		Style: `
			Y 70%
			W 30%
			H 30%
			BackgroundColor 128 128 128 128
		`,
	})
	s.GameContainer.AdoptChannel <- s.GroundWindow.This
	// Sub-window: stats
	err = s.StatsWindow.Setup(ui.ContainerConfig{
		Value: "Stats",
		Style: `
			X 30%
			W 40%
			H 20%
			BackgroundColor 128 0 0 128
		`,
	})
	s.GameContainer.AdoptChannel <- s.StatsWindow.This
	s.StatsWindow.SetHidden(true)
	// Sub-window: state
	err = s.StateWindow.Setup(ui.ContainerConfig{
		Value: "State",
		Style: `
			X 30%
			Y 80%
			W 40%
			H 20%
			BackgroundColor 128 128 0 128
		`,
	})
	s.GameContainer.AdoptChannel <- s.StateWindow.This
	s.StateWindow.SetHidden(true)
	//
	//go s.Client.LoopCmd()
	go s.Loop()
	return
}

// Close our Game state.
func (s *Game) Close() {
	s.MapWindow.Destroy()
	s.StateWindow.Destroy()
	s.StatsWindow.Destroy()
	s.GroundWindow.Destroy()
	s.InventoryWindow.Destroy()
	s.ChatWindow.Destroy()
}

// Loop is our loop for managing network activity and beyond.
func (s *Game) Loop() {
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

// HandleNet handles the network code for our Game state.
func (s *Game) HandleNet(cmd network.Command) bool {
	switch c := cmd.(type) {
	case network.CommandAnimation:
		s.Client.DataManager.HandleAnimationCommand(c)
	case network.CommandMap:
		s.world.HandleMapCommand(c)
	case network.CommandObject:
		s.world.HandleObjectCommand(c)
	case network.CommandTile:
		s.world.HandleTileCommand(c)
	default:
		s.Client.Log.Printf("Server sent a Command %+v\n", c)
	}
	return false
}
