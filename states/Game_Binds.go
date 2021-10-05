package states

import (
	"os"
	"strings"

	"github.com/chimera-rpg/go-client/binds"
	"github.com/chimera-rpg/go-client/ui"
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-common/network"
)

var (
	defaultNorth1 = binds.KeyGroup{
		Keys:    []uint8{107},
		Pressed: true,
	}
	defaultNorth2 = binds.KeyGroup{
		Keys:    []uint8{82},
		Pressed: true,
	}
	defaultNorthRun1 = binds.KeyGroup{
		Keys:    []uint8{107},
		Pressed: true,
		Repeat:  true,
	}
	defaultNorthRun2 = binds.KeyGroup{
		Keys:    []uint8{82},
		Pressed: true,
		Repeat:  true,
	}
	defaultSouth1 = binds.KeyGroup{
		Keys:    []uint8{106},
		Pressed: true,
	}
	defaultSouth2 = binds.KeyGroup{
		Keys:    []uint8{81},
		Pressed: true,
	}
	defaultSouthRun1 = binds.KeyGroup{
		Keys:    []uint8{106},
		Pressed: true,
		Repeat:  true,
	}
	defaultSouthRun2 = binds.KeyGroup{
		Keys:    []uint8{81},
		Pressed: true,
		Repeat:  true,
	}
	defaultWest1 = binds.KeyGroup{
		Keys:    []uint8{104},
		Pressed: true,
	}
	defaultWest2 = binds.KeyGroup{
		Keys:    []uint8{80},
		Pressed: true,
	}
	defaultWestRun1 = binds.KeyGroup{
		Keys:    []uint8{104},
		Pressed: true,
		Repeat:  true,
	}
	defaultWestRun2 = binds.KeyGroup{
		Keys:    []uint8{80},
		Pressed: true,
		Repeat:  true,
	}
	defaultEast1 = binds.KeyGroup{
		Keys:    []uint8{108},
		Pressed: true,
	}
	defaultEast2 = binds.KeyGroup{
		Keys:    []uint8{79},
		Pressed: true,
	}
	defaultEastRun1 = binds.KeyGroup{
		Keys:    []uint8{108},
		Pressed: true,
		Repeat:  true,
	}
	defaultEastRun2 = binds.KeyGroup{
		Keys:    []uint8{79},
		Pressed: true,
		Repeat:  true,
	}
	defaultFocusChat = binds.KeyGroup{
		Keys:    []uint8{13},
		Pressed: true,
	}
)

func (s *Game) SetupBinds() {
	// This isn't the right place for this.
	if s.Client.DataManager.Config.Game.CommandPrefix == "" {
		s.Client.DataManager.Config.Game.CommandPrefix = "/"
	}

	// Set up bindings.
	s.bindings = &s.Client.DataManager.Config.Game.Bindings
	s.bindings.Init()
	s.bindings.SetFunction("north", func(i ...interface{}) {
		s.Client.Send(network.CommandCmd{
			Cmd: network.North,
		})
	})
	s.bindings.SetFunction("north run", func(i ...interface{}) {
		s.Client.Send(network.CommandCmd{
			Cmd: network.North,
		})
	})
	s.bindings.SetFunction("south", func(i ...interface{}) {
		s.Client.Send(network.CommandCmd{
			Cmd: network.South,
		})
	})
	s.bindings.SetFunction("south run", func(i ...interface{}) {
		s.Client.Send(network.CommandCmd{
			Cmd: network.South,
		})
	})
	s.bindings.SetFunction("east", func(i ...interface{}) {
		s.Client.Send(network.CommandCmd{
			Cmd: network.East,
		})
	})
	s.bindings.SetFunction("east run", func(i ...interface{}) {
		s.Client.Send(network.CommandCmd{
			Cmd: network.East,
		})
	})
	s.bindings.SetFunction("west", func(i ...interface{}) {
		s.Client.Send(network.CommandCmd{
			Cmd: network.West,
		})
	})
	s.bindings.SetFunction("west run", func(i ...interface{}) {
		s.Client.Send(network.CommandCmd{
			Cmd: network.West,
		})
	})
	s.bindings.SetFunction("quit", func(i ...interface{}) {
		os.Exit(0)
	})
	s.bindings.SetFunction("disconnect", func(i ...interface{}) {
		go func() {
			s.inputChan <- DisconnectEvent{}
		}()
	})
	s.bindings.SetFunction("say", func(i ...interface{}) {
		str := ""
		if len(i) > 0 {
			switch v := i[0].(type) {
			case string:
				str = v
			case []string:
				str = strings.Join(v, " ")
			default:
				s.Print("say failed")
				return
			}
		}
		if str == "" {
			s.Print("say what?")
			return
		}
		s.Client.Send(network.CommandMessage{
			Type: network.PCMessage,
			Body: str,
		})
	})
	s.bindings.SetFunction("chat", func(i ...interface{}) {
		str := ""
		if len(i) > 0 {
			switch v := i[0].(type) {
			case string:
				str = v
			case []string:
				str = strings.Join(v, " ")
			default:
				s.Print("chat failed")
				return
			}
		}
		if str == "" {
			s.Print("chat what?")
			return
		}
		s.Client.Send(network.CommandMessage{
			Type: network.ChatMessage,
			Body: str,
		})
	})

	s.bindings.SetFunction("squeeze", func(i ...interface{}) {
		s.Client.Send(network.CommandStatus{
			Type:   cdata.SqueezingStatus,
			Active: true,
		})
	})

	s.bindings.SetFunction("crouch", func(i ...interface{}) {
		s.Client.Send(network.CommandStatus{
			Type:   cdata.CrouchingStatus,
			Active: true,
		})
	})

	s.bindings.SetFunction("focus chat", func(i ...interface{}) {
		s.ChatInput.GetUpdateChannel() <- ui.UpdateFocus{}
	})
	if !s.bindings.HasKeygroupsForName("north") {
		s.bindings.AddKeygroup("north", defaultNorth1)
		s.bindings.AddKeygroup("north", defaultNorth2)
	}
	if !s.bindings.HasKeygroupsForName("north run") {
		s.bindings.AddKeygroup("north run", defaultNorthRun1)
		s.bindings.AddKeygroup("north run", defaultNorthRun2)
	}
	if !s.bindings.HasKeygroupsForName("south") {
		s.bindings.AddKeygroup("south", defaultSouth1)
		s.bindings.AddKeygroup("south", defaultSouth2)
	}
	if !s.bindings.HasKeygroupsForName("south run") {
		s.bindings.AddKeygroup("south run", defaultSouthRun1)
		s.bindings.AddKeygroup("south run", defaultSouthRun2)
	}
	if !s.bindings.HasKeygroupsForName("west") {
		s.bindings.AddKeygroup("west", defaultWest1)
		s.bindings.AddKeygroup("west", defaultWest2)
	}
	if !s.bindings.HasKeygroupsForName("west run") {
		s.bindings.AddKeygroup("west run", defaultWestRun1)
		s.bindings.AddKeygroup("west run", defaultWestRun2)
	}
	if !s.bindings.HasKeygroupsForName("east") {
		s.bindings.AddKeygroup("east", defaultEast1)
		s.bindings.AddKeygroup("east", defaultEast2)
	}
	if !s.bindings.HasKeygroupsForName("east run") {
		s.bindings.AddKeygroup("east run", defaultEastRun1)
		s.bindings.AddKeygroup("east run", defaultEastRun2)
	}

	if !s.bindings.HasKeygroupsForName("focus chat") {
		s.bindings.AddKeygroup("focus chat", defaultFocusChat)
	}
}
