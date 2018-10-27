package Client

import (
  "github.com/veandco/go-sdl2/sdl"
  "github.com/veandco/go-sdl2/ttf"
  "client/UI"
  "common/Net"
  "fmt"
  "path"
  "os"
  "log"
)

type Client struct {
  RootWindow UI.Window
  DefaultFont *ttf.Font
  Net.Connection
  LogHistory []string
  State StateI
  DataRoot string
  Log *log.Logger
  isRunning bool
  RenderChannel chan struct{}
  StateChannel chan StateMessage
}

func NewClient() (c *Client, e error) {
  c = &Client{}
  e = c.Setup()
  return
}

func (c *Client) Setup() (err error) {
  c.Log = log.New(os.Stdout, "Client: ", log.Lshortfile)
  c.DataRoot = path.Join("share", "chimera", "client")
  err = c.RootWindow.Setup(UI.WindowConfig{
    Title: "Chimera",
    Dimensions: sdl.Rect{
      sdl.WINDOWPOS_UNDEFINED,
      sdl.WINDOWPOS_UNDEFINED,
      1280,
      720,
    },
    RenderFunc: func(w *UI.Window) {
      w.Renderer.SetDrawColor(0, 0, 0, 255)
      w.Renderer.Clear()
      w.Renderer.SetDrawColor(255, 0, 255, 255)
      w.Renderer.DrawPoint(150, 300)
      w.Renderer.DrawLine(0, 0, 200, 200)
    },
  })
  if err != nil {
    return
  }
  if c.DefaultFont, err = ttf.OpenFont(path.Join(c.DataRoot, "fonts", "DefaultFont.ttf"), 12); err != nil {
    return
  }
  Net.RegisterCommands()

  c.RenderChannel = make(chan struct{})
  c.StateChannel = make(chan StateMessage)

  // Render the initial window
  c.RootWindow.Render()

  c.isRunning = true
  return
}

func (c *Client) Destroy() {
  c.isRunning = false
  c.Close()
  c.State.Close()
  c.RootWindow.Destroy()
}

func (c *Client) Print(format string, a ...interface{}) {
  c.Log.Printf(format, a...)
  c.LogHistory = append(c.LogHistory, fmt.Sprintf(format, a...))
}

func (c *Client) SetState(state StateI, v interface{}) {
  if c.State != nil {
    c.State.Close()
  }
  state.SetClient(c)
  c.State = state
  next, nextArgs, err := c.State.Init(v)
  if err != nil {
    c.Log.Print(err)
  }
  if next != nil {
    c.SetState(next, nextArgs)
  }
}

func (c *Client) Render() {
  c.RootWindow.RenderMutex.Lock()
  c.State.HandleRender()
  c.RootWindow.Render()
  c.RootWindow.RenderMutex.Unlock()
}

func (c *Client) ChannelLoop() {
  for c.isRunning {
    select {
    case <- c.RenderChannel:
      c.Render()
    case msg := <- c.StateChannel:
      c.SetState(msg.State, msg.Args)
    }
  }
}

func (c *Client) IsRunning() bool {
  return c.isRunning
}
