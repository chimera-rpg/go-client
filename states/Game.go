package states

import (
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/ui"
)

// Game is our live Game state, used once the user has connected to the server
// and joined as a player character.
type Game struct {
	client.State
	ChatWindow      ui.Container
	MapWindow       ui.Container
	InventoryWindow ui.Container
	GroundWindow    ui.Container
	StatsWindow     ui.Container
	StateWindow     ui.Container
}

// Init our Game state.
func (s *Game) Init(t interface{}) (state client.StateI, nextArgs interface{}, err error) {
	s.Client.Log.Print("Game State")
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
	s.Client.RootWindow.AdoptChannel <- s.MapWindow.This
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
	s.Client.RootWindow.AdoptChannel <- s.ChatWindow.This
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
	s.Client.RootWindow.AdoptChannel <- s.InventoryWindow.This
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
	s.Client.RootWindow.AdoptChannel <- s.GroundWindow.This
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
	s.Client.RootWindow.AdoptChannel <- s.StatsWindow.This
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
	s.Client.RootWindow.AdoptChannel <- s.StateWindow.This
	s.StateWindow.SetHidden(true)
	//
	//go s.Client.LoopCmd()
	go s.HandleNet()
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

// HandleNet handles the network code for our Game state.
func (s *Game) HandleNet() {
	for s.Client.IsRunning() {
		select {
		case cmd := <-s.Client.CmdChan:
			s.Client.Log.Printf("cmd! %d", cmd.GetType())
		case <-s.Client.ClosedChan:
			s.Client.Log.Print("Lost connection to server.")
			s.Client.StateChannel <- client.StateMessage{State: &List{}, Args: nil}
		}
	}
}
