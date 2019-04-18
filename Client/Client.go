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
	RootWindow  *UI.Window
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
	c.DataRoot = path.Join("share", "chimera", "client")
	return
}

func (c *Client) Setup(inst *UI.Instance) (err error) {
	c.Log = log.New(os.Stdout, "Client: ", log.Lshortfile)

	/*err = UI.RootWindow.Setup(UI.WindowConfig{
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
		Context: &UI.Context{},
	})
	if err != nil {
		return
	}*/
	c.RootWindow = &inst.RootWindow

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

func (c *Client) ChannelLoop() {
	for c.isRunning {
		select {
		/*case <-c.RenderChannel:
		c.Refresh()*/
		case msg := <-c.StateChannel:
			if c.isRunning {
				c.SetState(msg.State, msg.Args)
			}
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
