package main

import (
  "runtime/debug"
  "github.com/chimera-rpg/go-client/Client"
  "github.com/chimera-rpg/go-client/States"
  "github.com/veandco/go-sdl2/sdl"
  "github.com/veandco/go-sdl2/ttf"
  "log"
  "fmt"
  "flag"
)

func showWindow(flags uint32, format string, a ...interface{}) {
  buttons := []sdl.MessageBoxButtonData{
    {sdl.MESSAGEBOX_BUTTON_RETURNKEY_DEFAULT, 1, "OH NO"},
  }

  messageboxdata := sdl.MessageBoxData{
    flags,
    nil,
    "Chimera",
    fmt.Sprintf(format, a...),
    buttons,
    nil,
  }

  sdl.ShowMessageBox(&messageboxdata)
}

func showError(format string, a ...interface{}) {
  showWindow(sdl.MESSAGEBOX_ERROR, format, a)
}
func showWarning(format string, a ...interface{}) {
  showWindow(sdl.MESSAGEBOX_WARNING, format, a)
}
func showInfo(format string, a ...interface{}) {
  showWindow(sdl.MESSAGEBOX_INFORMATION, format, a)
}

func main() {
  defer func() {
    if r := recover(); r != nil {
      showError("%v", r.(error).Error())
      debug.PrintStack()
    }
  }()
  log.Print("Starting Chimera client (golang)")

  // Initialize SDL
  if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
    panic(err)
  }
  defer sdl.Quit()
  // Initialize TTF
  if err := ttf.Init(); err != nil {
    panic(err)
  }

  client, err := Client.NewClient()
  defer client.Destroy()
  if err != nil {
    showError("%s", err)
    return
  }

  // Start the client's channel listening loop as a coroutine
  go client.ChannelLoop()

  netPtr := flag.String("connect", "", "SERVER:PORT")
  flag.Parse()

  // Automatically attempt to connect if the server flag was passed
  if (len(*netPtr) > 0) {
    client.StateChannel <- Client.StateMessage{&States.Handshake{}, *netPtr}
  } else {
    client.StateChannel <- Client.StateMessage{&States.List{}, nil}
  }

  var empty struct{}

  // Run the SDL loop in the main thread, as calling WaitEvent() in a coroutine causes crashes on Mac OS.
  running := true
  for running {
    event := sdl.WaitEvent()
    client.RenderChannel <- empty
    switch t := event.(type) {
    case *sdl.QuitEvent:
      running = false
    case *sdl.WindowEvent:
      if t.Event == sdl.WINDOWEVENT_RESIZED {
        client.RootWindow.RenderMutex.Lock()
        client.RootWindow.Resize(t.WindowID, t.Data1, t.Data2)
        client.RootWindow.RenderMutex.Unlock()
      } else if t.Event == sdl.WINDOWEVENT_CLOSE {
        running = false
      } else if t.Event == sdl.WINDOWEVENT_EXPOSED {
        client.Render()
      }
    default:
      client.State.HandleEvent(&event)
      client.Refresh()
    }
  }

  log.Print("Sayonara!")
}
