package Client

import (
	"fmt"
	"github.com/chimera-rpg/go-client/UI"
	"github.com/chimera-rpg/go-common/Net"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"log"
	"os"
	"path"
)

type Client struct {
	RootWindow  UI.Window
	DefaultFont *ttf.Font
	Net.Connection
	LogHistory    []string
	State         StateI
	DataRoot      string
	Log           *log.Logger
	isRunning     bool
	RenderChannel chan struct{}
	StateChannel  chan StateMessage
	EventChannel  chan sdl.Event
}

func NewClient() (c *Client, e error) {
	c = &Client{}
	e = c.Setup()
	return
}

func (c *Client) Setup() (err error) {
	c.Log = log.New(os.Stdout, "Client: ", log.Lshortfile)
	c.DataRoot = path.Join("share", "chimera", "client")

	context := UI.Context{}
	if context.Font, err = ttf.OpenFont(path.Join(c.DataRoot, "fonts", "DefaultFont.ttf"), 12); err != nil {
		return
	}

	err = c.RootWindow.Setup(UI.WindowConfig{
		Value: "Chimera",
		Style: UI.Style{
			X: UI.Number{Value: 0},
			Y: UI.Number{Value: 0},
			W: UI.Number{Value: 1280},
			H: UI.Number{Value: 720},
		},
		RenderFunc: func(w *UI.Window) {
			w.Context.Renderer.SetDrawColor(128, 196, 128, 255)
			w.Context.Renderer.Clear()
		},
		Context: &context,
	})
	if err != nil {
		return
	}

	Net.RegisterCommands()

	c.RenderChannel = make(chan struct{})
	c.StateChannel = make(chan StateMessage)
	c.EventChannel = make(chan sdl.Event)

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
	c.RootWindow.Render()
	c.RootWindow.RenderMutex.Unlock()
}

func (c *Client) Refresh() {
	if c.RootWindow.HasDirt() {
		c.Render()
	}
}

func (c *Client) ChannelLoop() {
	for c.isRunning {
		select {
		case <-c.RenderChannel:
			c.Refresh()
		case msg := <-c.StateChannel:
			c.SetState(msg.State, msg.Args)
		}
	}
}

func (c *Client) GetPNGData(file string) (data []byte) {
	reader, err := os.Open(path.Join(c.DataRoot, file))
	if err != nil {
		panic(err)
	}
	info, err := reader.Stat()
	if err != nil {
		panic(err)
	}
	data = make([]byte, info.Size())
	_, err = reader.Read(data)
	if err != nil {
		panic(err)
	}
	return
}

func (c *Client) IsRunning() bool {
	return c.isRunning
}
