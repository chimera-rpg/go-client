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
	CurrentServer    string
	DataManager      *data.Manager
	RootWindow       *ui.Window
	LogHistory       []string
	States           []StateI
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
	c.CloseState(c.State())
}

// Print provides an interface to Log that is instantiated to the Client itself.
func (c *Client) Print(format string, a ...interface{}) {
	c.Log.Printf(format, a...)
	c.LogHistory = append(c.LogHistory, fmt.Sprintf(format, a...))
}

// ReplaceState sets the current state to the provided one, optionally passing v
// to the next state. Calls Close() on the current state.
func (c *Client) ReplaceState(state StateI, v interface{}) {
	if c.State() != nil {
		c.CloseState(c.State())
	}
	c.SetupState(state)
	if c.State() != nil {
		c.States[len(c.States)-1] = state
	} else {
		c.States = append(c.States, state)
	}
	c.State().SetRunning(true)
	next, nextArgs, err := c.State().Init(v)
	if err != nil {
		c.Log.Error(err)
	}
	if next != nil {
		c.ReplaceState(next, nextArgs)
	}
}

func (c *Client) SetupState(state StateI) {
	state.SetClient(c)
	state.CreateChannels()
}

func (c *Client) CloseState(state StateI) {
	state.Close()
	select {
	case state.GetCloseChannel() <- true:
	default:
	}
}

func (c *Client) PushState(state StateI, v interface{}) {
	if c.State() != nil {
		c.State().Leave()
		c.State().SetRunning(false)
	}
	c.SetupState(state)
	c.States = append(c.States, state)
	c.State().SetRunning(true)
	next, nextArgs, err := c.State().Init(v)
	if err != nil {
		c.Log.Error(err)
	}
	if next != nil {
		c.ReplaceState(next, nextArgs)
	}
}

func (c *Client) PopState(enterArgs ...interface{}) {
	if c.State() != nil {
		c.CloseState(c.State())
		c.States = c.States[:len(c.States)-1]
	}
	if c.State() != nil {
		c.State().Enter(enterArgs)
		c.State().SetRunning(true)
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
				if msg.Push {
					c.PushState(msg.State, msg.Args)
				} else if msg.Pop {
					c.PopState(msg.Args)
				} else if msg.PopToTop {
					if len(c.States) > 1 {
						for i := len(c.States) - 1; i >= 1; i-- {
							c.CloseState(c.State())
							c.States = c.States[:len(c.States)-1]
						}
						c.State().SetRunning(true)
						c.State().Enter(msg.Args)
					}
				} else {
					c.ReplaceState(msg.State, msg.Args)
				}
			}
		}
	}
}

func (c *Client) State() StateI {
	if len(c.States) > 0 {
		return c.States[len(c.States)-1]
	}
	return nil
}

// IsRunning returns whether the client is running or not.
func (c *Client) IsRunning() bool {
	return c.isRunning
}
