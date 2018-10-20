package States

import (
  "client/Client"
  "client/UI"
  "github.com/veandco/go-sdl2/sdl"
)

type List struct {
  Client.State
  ServersWindow UI.Window
}

func (s *List) Init(t interface{}) (state Client.StateI, nextArgs interface{}, err error) {
  s.Client.RootWindow.RenderMutex.Lock()
  defer s.Client.RootWindow.RenderMutex.Unlock()
  err = s.ServersWindow.Setup(UI.WindowConfig{
    Title: "Server List",
    ParentDimensions: sdl.Rect{
      10,
      10,
      80,
      80,
    },
    Window: &s.Client.RootWindow,
    RenderFunc: func(w *UI.Window) {
      w.Renderer.SetDrawColor(0, 0, 0, 255)
      w.Renderer.Clear()
    },
  })

  el := UI.Element(&UI.Text{
    BaseElement: UI.BaseElement{
      Color: sdl.Color{255, 255, 255, 255},
      Value: "Please choose a server:",
      Position: sdl.Point{100, 100},
    },
    Font: s.Client.DefaultFont,
  })

  s.ServersWindow.AddElement(&el)

  s.Client.Print("Please choose a server: ")
  return
}

func (s *List) Close() {
  s.ServersWindow.Destroy()
}
