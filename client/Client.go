package client

import (
	"fmt"

	"github.com/chimera-rpg/go-client/data"
	"github.com/chimera-rpg/go-client/ui"
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-common/network"
	"github.com/sirupsen/logrus"
)

// Client is the main handler of state, network transmission, and otherwise.
type Client struct {
	network.Connection
	DataManager      *data.Manager
	RootWindow       *ui.Window
	MessageHistory   []Message
	LogHistory       []string
	State            StateI
	Log              *logrus.Logger
	isRunning        bool
	RenderChannel    chan struct{}
	StateChannel     chan StateMessage
	AnimationsConfig cdata.AnimationsConfig
}

// Setup sets up a Client's base data structures for use.
func (c *Client) Setup(dataManager *data.Manager, inst *ui.Instance, l *logrus.Logger) (err error) {
	c.Log = l

	c.RootWindow = &inst.RootWindow
	c.DataManager = dataManager
	c.DataManager.Conn = &c.Connection

	network.RegisterCommands()

	c.RenderChannel = make(chan struct{})
	c.StateChannel = make(chan StateMessage)

	c.isRunning = true
	return
}

// Destroy cleans up the client and its last sate.
func (c *Client) Destroy() {
	c.isRunning = false
	c.Close()
	c.State.Close()
}

// Print provides an interface to Log that is instantiated to the Client itself.
func (c *Client) Print(format string, a ...interface{}) {
	c.Log.Printf(format, a...)
	c.LogHistory = append(c.LogHistory, fmt.Sprintf(format, a...))
}

// SetState sets the current state to the provided one, optionally passing v
// to the next state. Calls Close() on the current state.
func (c *Client) SetState(state StateI, v interface{}) {
	if c.State != nil {
		c.State.Close()
		select {
		case c.State.GetCloseChannel() <- true:
		default:
		}
	}
	state.SetClient(c)
	state.CreateChannels()
	c.State = state
	next, nextArgs, err := c.State.Init(v)
	if err != nil {
		c.Log.Error(err)
	}
	if next != nil {
		c.SetState(next, nextArgs)
	}
}

// ChannelLoop is the client's go routine for listening to and responding to
// its channels.
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

// IsRunning returns whether the client is running or not.
func (c *Client) IsRunning() bool {
	return c.isRunning
}
