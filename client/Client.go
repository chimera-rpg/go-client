package client

import (
	"fmt"

	"github.com/chimera-rpg/go-client/audio"
	"github.com/chimera-rpg/go-client/config"
	"github.com/chimera-rpg/go-client/data"
	"github.com/chimera-rpg/go-client/ui"
	cdata "github.com/chimera-rpg/go-server/data"
	"github.com/chimera-rpg/go-server/network"
	"github.com/sirupsen/logrus"
)

// Client is the main handler of state, network transmission, and otherwise.
type Client struct {
	network.Connection
	CurrentServer    string
	DataManager      *data.Manager
	Flags            config.Flags
	RootWindow       *ui.Window
	UI               *ui.Instance
	Audio            *audio.Instance
	LogHistory       []string
	States           []StateI
	Log              *logrus.Logger
	isRunning        bool
	RenderChannel    chan struct{}
	StateChannel     chan StateMessage
	AnimationsConfig clientAnimationsConfig
	// TODO: Probably move this elsewhere.
	TypeHints map[uint32]string
	Slots     map[uint32]string
}

// Setup sets up a Client's base data structures for use.
func (c *Client) Setup(dataManager *data.Manager, inst *ui.Instance, aud *audio.Instance, l *logrus.Logger) (err error) {
	c.Log = l

	c.RootWindow = &inst.RootWindow
	c.Audio = aud
	c.UI = inst
	c.DataManager = dataManager
	c.DataManager.Conn = &c.Connection

	network.RegisterCommands()

	c.RenderChannel = make(chan struct{})
	c.StateChannel = make(chan StateMessage)

	c.TypeHints = make(map[uint32]string)
	c.Slots = make(map[uint32]string)

	c.Flags.Parse()

	c.isRunning = true
	return
}

// Destroy cleans up the client, its last state, and saves out the config.
func (c *Client) Destroy() {
	c.isRunning = false
	c.Close()
	c.CloseState(c.State())
	c.DataManager.Config.Write()
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

// LoadAnimationsConfig converts the map-based cdata.AnimationsConfig to a slice-based one that is more efficient for the client to constantly access.
func (c *Client) LoadAnimationsConfig(conf cdata.AnimationsConfig) {
	c.AnimationsConfig.TileWidth = int(conf.TileWidth)
	c.AnimationsConfig.TileHeight = int(conf.TileHeight)
	c.AnimationsConfig.YStep.X = int(conf.YStep.X)
	c.AnimationsConfig.YStep.Y = int(conf.YStep.Y)
	for t, a := range conf.Adjustments {
		c.AnimationsConfig.Adjustments = append(c.AnimationsConfig.Adjustments, archetypeAnimationAdjustment{
			Type: t,
			X:    int(a.X),
			Y:    int(a.Y),
		})
	}
}

type clientAnimationsConfig struct {
	TileWidth  int
	TileHeight int
	YStep      struct {
		X int
		Y int
	}
	Adjustments []archetypeAnimationAdjustment
}

type archetypeAnimationAdjustment struct {
	Type cdata.ArchetypeType
	X    int
	Y    int
}

func (c *clientAnimationsConfig) GetAdjustment(t cdata.ArchetypeType) (archetypeAnimationAdjustment, bool) {
	for _, a := range c.Adjustments {
		if a.Type == t {
			return a, true
		}
	}
	return archetypeAnimationAdjustment{}, false
}
