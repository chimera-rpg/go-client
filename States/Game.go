package States

import (
	"github.com/chimera-rpg/go-client/Client"
	"github.com/chimera-rpg/go-client/UI"
	"github.com/veandco/go-sdl2/sdl"
)

type Game struct {
	Client.State
	ChatWindow      UI.Window
	MapWindow       UI.Window
	InventoryWindow UI.Window
	GroundWindow    UI.Window
	StatsWindow     UI.Window
	StateWindow     UI.Window
}

func (s *Game) Init(t interface{}) (state Client.StateI, nextArgs interface{}, err error) {
	s.Client.Log.Print("Game State")
	// Sub-window: map
	err = s.MapWindow.Setup(UI.WindowConfig{
		Value: "Map",
		Style: `
			X 50%
			Y 50%
			W 100%
			H 100%
			Origin CenterX CenterY
		`,
		Parent: s.Client.RootWindow,
		RenderFunc: func(w *UI.Window) {
			w.Context.Renderer.SetDrawColor(0, 128, 0, 128)
			w.Context.Renderer.Clear()
			//
			for x := 0; x < 12; x++ {
				for y := 0; y < 12; y++ {
					w.Context.Renderer.FillRect(&sdl.Rect{
						16 * 3 * int32(x), 16 * 3 * int32(y), 16 * 3, 16 * 3,
					})
				}
			}
		},
	})
	// Sub-window: chat
	err = s.ChatWindow.Setup(UI.WindowConfig{
		Value: "Chat",
		Style: `
			X 8
			Y 8
			W 70%
			H 20%
		`,
		Parent: s.Client.RootWindow,
		RenderFunc: func(w *UI.Window) {
			w.Context.Renderer.SetDrawColor(255, 0, 0, 128)
			w.Context.Renderer.Clear()
		},
	})
	// Sub-window: inventory
	err = s.InventoryWindow.Setup(UI.WindowConfig{
		Value: "Inventory",
		Style: `
			X 50%
			Y 50%
			W 50%
			H 80%
			Origin CenterX CenterY
		`,
		Parent: s.Client.RootWindow,
		RenderFunc: func(w *UI.Window) {
			w.Context.Renderer.SetDrawColor(0, 255, 0, 255)
			w.Context.Renderer.Clear()
		},
	})
	s.InventoryWindow.SetHidden(true)
	// Sub-window: ground
	err = s.GroundWindow.Setup(UI.WindowConfig{
		Value: "Ground",
		Style: `
			Y 70%
			W 30%
			H 30%
		`,
		Parent: s.Client.RootWindow,
		RenderFunc: func(w *UI.Window) {
			w.Context.Renderer.SetDrawColor(255, 0, 0, 128)
			w.Context.Renderer.Clear()
		},
	})
	// Sub-window: stats
	err = s.StatsWindow.Setup(UI.WindowConfig{
		Value: "Stats",
		Style: `
			X 30%
			W 40%
			H 20%
		`,
		Parent: s.Client.RootWindow,
		RenderFunc: func(w *UI.Window) {
			w.Context.Renderer.SetDrawColor(0, 0, 255, 255)
			w.Context.Renderer.Clear()
		},
	})
	s.StatsWindow.SetHidden(true)
	// Sub-window: state
	err = s.StateWindow.Setup(UI.WindowConfig{
		Value: "State",
		Style: `
			X 30%
			Y 80%
			W 40%
			H 20%
		`,
		Parent: s.Client.RootWindow,
		RenderFunc: func(w *UI.Window) {
			w.Context.Renderer.SetDrawColor(0, 0, 255, 255)
			w.Context.Renderer.Clear()
		},
	})
	s.StateWindow.SetHidden(true)
	//
	//go s.Client.LoopCmd()
	go s.HandleNet()
	return
}

func (s *Game) Close() {
	s.MapWindow.Destroy()
	s.StateWindow.Destroy()
	s.StatsWindow.Destroy()
	s.GroundWindow.Destroy()
	s.InventoryWindow.Destroy()
	s.ChatWindow.Destroy()
}

func (s *Game) HandleNet() {
	for s.Client.IsRunning() {
		select {
		case cmd := <-s.Client.CmdChan:
			s.Client.Log.Printf("cmd! %d", cmd.GetType())
		case <-s.Client.ClosedChan:
			s.Client.Log.Print("Lost connection to server.")
			s.Client.StateChannel <- Client.StateMessage{&List{}, nil}
		}
	}
	/*defer func() {
	    if r := recover(); r != nil {
	      s.Client.Log.Print("Guess we done.")
	    }
	  }()
	  for s.NetListening {
	    var cmd Net.Command
	    s.Client.Log.Print("Pre cmd")
	    s.Client.Receive(&cmd)
	    s.Client.Log.Print("Post cmd")
	  }*/
}
