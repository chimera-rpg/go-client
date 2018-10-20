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
  s.Client.RootWindow.RenderMutex.Lock()
  defer s.Client.RootWindow.RenderMutex.Unlock()
  // Sub-window: chat
  err = s.ChatWindow.Setup(UI.WindowConfig{
    Title: "Chat",
    ParentDimensions: sdl.Rect{
      70,
      0,
      30,
      100,
    },
    Window: &s.Client.RootWindow,
    RenderFunc: func(w *UI.Window) {
      w.Renderer.SetDrawColor(255, 0, 0, 255)
      w.Renderer.Clear()
    },
  })
  // Sub-window: inventory
  err = s.InventoryWindow.Setup(UI.WindowConfig{
    Title: "Inventory",
    ParentDimensions: sdl.Rect{
      0,
      0,
      30,
      70,
    },
    Window: &s.Client.RootWindow,
    RenderFunc: func(w *UI.Window) {
      w.Renderer.SetDrawColor(0, 255, 0, 255)
      w.Renderer.Clear()
    },
  })
  // Sub-window: ground
  err = s.GroundWindow.Setup(UI.WindowConfig{
    Title: "Ground",
    ParentDimensions: sdl.Rect{
      0,
      70,
      30,
      30,
    },
    Window: &s.Client.RootWindow,
    RenderFunc: func(w *UI.Window) {
      w.Renderer.SetDrawColor(255, 0, 0, 255)
      w.Renderer.Clear()
    },
  })
  // Sub-window: stats
  err = s.StatsWindow.Setup(UI.WindowConfig{
    Title: "Stats",
    ParentDimensions: sdl.Rect{
      30,
      0,
      40,
      20,
    },
    Window: &s.Client.RootWindow,
    RenderFunc: func(w *UI.Window) {
      w.Renderer.SetDrawColor(0, 0, 255, 255)
      w.Renderer.Clear()
    },
  })
  // Sub-window: state
  err = s.StateWindow.Setup(UI.WindowConfig{
    Title: "State",
    ParentDimensions: sdl.Rect{
      30,
      80,
      40,
      20,
    },
    Window: &s.Client.RootWindow,
    RenderFunc: func(w *UI.Window) {
      w.Renderer.SetDrawColor(0, 0, 255, 255)
      w.Renderer.Clear()
    },
  })
  // Sub-window: map
  err = s.MapWindow.Setup(UI.WindowConfig{
    Title: "Map",
    ParentDimensions: sdl.Rect{
      30,
      20,
      40,
      60,
    },
    Window: &s.Client.RootWindow,
    RenderFunc: func(w *UI.Window) {
      w.Renderer.SetDrawColor(0, 0, 0, 255)
      w.Renderer.Clear()
      w.Renderer.SetDrawColor(255, 0, 255, 255)
      w.Renderer.DrawPoint(150, 300)
      w.Renderer.DrawLine(0, 0, 200, 200)
    },
  })
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
