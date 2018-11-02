package States

import (
  "client/Client"
  "client/UI"
  "github.com/veandco/go-sdl2/sdl"
)

type Game struct {
  Client.State
  ChatWindow UI.Window
  MapWindow UI.Window
  InventoryWindow UI.Window
  GroundWindow UI.Window
  StatsWindow UI.Window
  StateWindow UI.Window
}

func (s *Game) Init(t interface{}) (state Client.StateI, nextArgs interface{}, err error) {
  s.Client.Log.Print("Game State")
  s.Client.RootWindow.RenderMutex.Lock()
  defer s.Client.RootWindow.RenderMutex.Unlock()
  // Sub-window: map
  err = s.MapWindow.Setup(UI.WindowConfig{
    Value: "Map",
    Style: UI.Style{
      X: UI.Number{
        Percentage: true,
        Value: 50,
      },
      Y: UI.Number{
        Percentage: true,
        Value: 50,
      },
      W: UI.Number{
        Percentage: true,
        Value: 100,
      },
      H: UI.Number{
        Percentage: true,
        Value: 100,
      },
      Origin: UI.ORIGIN_CENTERX | UI.ORIGIN_CENTERY,
    },
    Parent: &s.Client.RootWindow,
    RenderFunc: func(w *UI.Window) {
      w.Context.Renderer.SetDrawColor(0, 128, 0, 128)
      w.Context.Renderer.Clear()
      //
      for x := 0; x < 12; x++ {
        for y := 0; y < 12; y ++ {
          w.Context.Renderer.FillRect(&sdl.Rect{
            16*3*int32(x), 16*3*int32(y), 16*3, 16*3,
          })
        }
      }
    },
  })
  // Sub-window: chat
  err = s.ChatWindow.Setup(UI.WindowConfig{
    Value: "Chat",
    Style: UI.Style{
      X: UI.Number{
        Value: 8,
      },
      Y: UI.Number{
        Value: 8,
      },
      W: UI.Number{
        Percentage: true,
        Value: 70,
      },
      H: UI.Number{
        Percentage: true,
        Value: 20,
      },
    },
    Parent: &s.Client.RootWindow,
    RenderFunc: func(w *UI.Window) {
      w.Context.Renderer.SetDrawColor(255, 0, 0, 128)
      w.Context.Renderer.Clear()
    },
  })
  // Sub-window: inventory
  err = s.InventoryWindow.Setup(UI.WindowConfig{
    Value: "Inventory",
    Style: UI.Style{
      X: UI.Number{
        Percentage: true,
        Value: 50,
      },
      Y: UI.Number{
        Percentage: true,
        Value: 50,
      },
      W: UI.Number{
        Percentage: true,
        Value: 50,
      },
      H: UI.Number{
        Percentage: true,
        Value: 80,
      },
      Origin: UI.ORIGIN_CENTERX | UI.ORIGIN_CENTERY,
    },
    Parent: &s.Client.RootWindow,
    RenderFunc: func(w *UI.Window) {
      w.Context.Renderer.SetDrawColor(0, 255, 0, 255)
      w.Context.Renderer.Clear()
    },
  })
  s.InventoryWindow.SetHidden(true)
  // Sub-window: ground
  err = s.GroundWindow.Setup(UI.WindowConfig{
    Value: "Ground",
    Style: UI.Style{
      X: UI.Number{
        Percentage: true,
      },
      Y: UI.Number{
        Percentage: true,
        Value: 70,
      },
      W: UI.Number{
        Percentage: true,
        Value: 30,
      },
      H: UI.Number{
        Percentage: true,
        Value: 30,
      },
    },
    Parent: &s.Client.RootWindow,
    RenderFunc: func(w *UI.Window) {
      w.Context.Renderer.SetDrawColor(255, 0, 0, 128)
      w.Context.Renderer.Clear()
    },
  })
  // Sub-window: stats
  err = s.StatsWindow.Setup(UI.WindowConfig{
    Value: "Stats",
    Style: UI.Style{
      X: UI.Number{
        Percentage: true,
        Value: 30,
      },
      Y: UI.Number{
        Percentage: true,
      },
      W: UI.Number{
        Percentage: true,
        Value: 40,
      },
      H: UI.Number{
        Percentage: true,
        Value: 20,
      },
    },
    Parent: &s.Client.RootWindow,
    RenderFunc: func(w *UI.Window) {
      w.Context.Renderer.SetDrawColor(0, 0, 255, 255)
      w.Context.Renderer.Clear()
    },
  })
  s.StatsWindow.SetHidden(true)
  // Sub-window: state
  err = s.StateWindow.Setup(UI.WindowConfig{
    Value: "State",
    Style: UI.Style{
      X: UI.Number{
        Percentage: true,
        Value: 30,
      },
      Y: UI.Number{
        Percentage: true,
        Value: 80,
      },
      W: UI.Number{
        Percentage: true,
        Value: 40,
      },
      H: UI.Number{
        Percentage: true,
        Value: 20,
      },
    },
    Parent: &s.Client.RootWindow,
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
    case cmd := <- s.Client.CmdChan:
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
