package states

import (
	"github.com/chimera-rpg/go-client/binds"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-common/network"
)

func (s *Game) SetupBinds() {
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

	s.bindings.SetFunction("focus chat", func(i ...interface{}) {
		s.ChatInput.GetUpdateChannel() <- ui.UpdateFocus{}
	})
	if !s.bindings.HasKeygroupsForName("north") {
		s.bindings.AddKeygroup("north", binds.KeyGroup{
			Keys:    []uint8{107},
			Pressed: true,
		})
		s.bindings.AddKeygroup("north", binds.KeyGroup{
			Keys:    []uint8{82},
			Pressed: true,
		})
	}
	if !s.bindings.HasKeygroupsForName("north run") {
		s.bindings.AddKeygroup("north run", binds.KeyGroup{
			Keys:    []uint8{107},
			Pressed: true,
			Repeat:  true,
		})
		s.bindings.AddKeygroup("north run", binds.KeyGroup{
			Keys:    []uint8{82},
			Pressed: true,
			Repeat:  true,
		})
	}
	if !s.bindings.HasKeygroupsForName("south") {
		s.bindings.AddKeygroup("south", binds.KeyGroup{
			Keys:    []uint8{106},
			Pressed: true,
		})
		s.bindings.AddKeygroup("south", binds.KeyGroup{
			Keys:    []uint8{81},
			Pressed: true,
		})
	}
	if !s.bindings.HasKeygroupsForName("south run") {
		s.bindings.AddKeygroup("south run", binds.KeyGroup{
			Keys:    []uint8{106},
			Pressed: true,
			Repeat:  true,
		})
		s.bindings.AddKeygroup("south run", binds.KeyGroup{
			Keys:    []uint8{81},
			Pressed: true,
			Repeat:  true,
		})
	}
	if !s.bindings.HasKeygroupsForName("west") {
		s.bindings.AddKeygroup("west", binds.KeyGroup{
			Keys:    []uint8{104},
			Pressed: true,
		})
		s.bindings.AddKeygroup("west", binds.KeyGroup{
			Keys:    []uint8{80},
			Pressed: true,
		})
	}
	if !s.bindings.HasKeygroupsForName("west run") {
		s.bindings.AddKeygroup("west run", binds.KeyGroup{
			Keys:    []uint8{104},
			Pressed: true,
			Repeat:  true,
		})
		s.bindings.AddKeygroup("west run", binds.KeyGroup{
			Keys:    []uint8{80},
			Pressed: true,
			Repeat:  true,
		})
	}
	if !s.bindings.HasKeygroupsForName("east") {
		s.bindings.AddKeygroup("east", binds.KeyGroup{
			Keys:    []uint8{108},
			Pressed: true,
		})
		s.bindings.AddKeygroup("east", binds.KeyGroup{
			Keys:    []uint8{79},
			Pressed: true,
		})
	}
	if !s.bindings.HasKeygroupsForName("east run") {
		s.bindings.AddKeygroup("east run", binds.KeyGroup{
			Keys:    []uint8{108},
			Pressed: true,
			Repeat:  true,
		})
		s.bindings.AddKeygroup("east run", binds.KeyGroup{
			Keys:    []uint8{79},
			Pressed: true,
			Repeat:  true,
		})
	}

	if !s.bindings.HasKeygroupsForName("focus chat") {
		s.bindings.AddKeygroup("focus chat", binds.KeyGroup{
			Keys:    []uint8{13},
			Pressed: true,
		})
	}
}
