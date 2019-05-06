package states

import (
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/ui"
)

// Game is our live Game state, used once the user has connected to the server
// and joined as a player character.
type Game struct {
	client.State
	ChatWindow      ui.Window
	MapWindow       ui.Window
	InventoryWindow ui.Window
	GroundWindow    ui.Window
	StatsWindow     ui.Window
	StateWindow     ui.Window
}

// Init our Game state.
func (s *Game) Init(t interface{}) (state client.StateI, nextArgs interface{}, err error) {
	s.Client.Log.Print("Game State")
	// Sub-window: map
	err = s.MapWindow.Setup(ui.WindowConfig{
		Value: "Map",
		Style: `
			X 50%
			Y 50%
			W 100%
			H 100%
			Origin CenterX CenterY
		`,
		Parent: s.Client.RootWindow,
	})
	// Sub-window: chat
	err = s.ChatWindow.Setup(ui.WindowConfig{
		Value: "Chat",
		Style: `
			X 8
			Y 8
			W 70%
			H 20%
		`,
		Parent: s.Client.RootWindow,
	})
	// Sub-window: inventory
	err = s.InventoryWindow.Setup(ui.WindowConfig{
		Value: "Inventory",
		Style: `
			X 50%
			Y 50%
			W 50%
			H 80%
			Origin CenterX CenterY
		`,
		Parent: s.Client.RootWindow,
	})
	s.InventoryWindow.SetHidden(true)
	// Sub-window: ground
	err = s.GroundWindow.Setup(ui.WindowConfig{
		Value: "Ground",
		Style: `
			Y 70%
			W 30%
			H 30%
		`,
		Parent: s.Client.RootWindow,
	})
	// Sub-window: stats
	err = s.StatsWindow.Setup(ui.WindowConfig{
		Value: "Stats",
		Style: `
			X 30%
			W 40%
			H 20%
		`,
		Parent: s.Client.RootWindow,
	})
	s.StatsWindow.SetHidden(true)
	// Sub-window: state
	err = s.StateWindow.Setup(ui.WindowConfig{
		Value: "State",
		Style: `
			X 30%
			Y 80%
			W 40%
			H 20%
		`,
		Parent: s.Client.RootWindow,
	})
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
	/*defer func() {
	    if r := recover(); r != nil {
	      s.Client.Log.Print("Guess we done.")
	    }
	  }()
	  for s.NetListening {
	    var cmd network.Command
	    s.Client.Log.Print("Pre cmd")
	    s.Client.Receive(&cmd)
	    s.Client.Log.Print("Post cmd")
	  }*/
}
