package game

import (
	"os"
	"strings"

	"github.com/chimera-rpg/go-client/binds"
	"github.com/chimera-rpg/go-client/ui"
	cdata "github.com/chimera-rpg/go-server/data"
	"github.com/chimera-rpg/go-server/network"
)

var (
	defaultClearCommands = binds.KeyGroup{
		Keys:    []uint8{96}, // ~
		Pressed: true,
	}
	defaultClearFocus = binds.KeyGroup{
		Keys:    []uint8{27}, // esc
		Pressed: false,
	}
	defaultNorth1 = binds.KeyGroup{
		Keys:    []uint8{107},
		Pressed: true,
	}
	defaultNorth2 = binds.KeyGroup{
		Keys:    []uint8{82},
		Pressed: true,
	}
	defaultNorthRun1 = binds.KeyGroup{
		Keys:     []uint8{107},
		Pressed:  true,
		Repeat:   true,
		OnRepeat: 1,
	}
	defaultNorthRun2 = binds.KeyGroup{
		Keys:     []uint8{82},
		Pressed:  true,
		Repeat:   true,
		OnRepeat: 1,
	}
	defaultNorthRunStop1 = binds.KeyGroup{
		Keys:    []uint8{107},
		Pressed: false,
	}
	defaultNorthRunStop2 = binds.KeyGroup{
		Keys:    []uint8{82},
		Pressed: false,
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
		Keys:     []uint8{106},
		Pressed:  true,
		Repeat:   true,
		OnRepeat: 1,
	}
	defaultSouthRun2 = binds.KeyGroup{
		Keys:     []uint8{81},
		Pressed:  true,
		Repeat:   true,
		OnRepeat: 1,
	}
	defaultSouthRunStop1 = binds.KeyGroup{
		Keys:    []uint8{106},
		Pressed: false,
	}
	defaultSouthRunStop2 = binds.KeyGroup{
		Keys:    []uint8{81},
		Pressed: false,
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
		Keys:     []uint8{104},
		Pressed:  true,
		Repeat:   true,
		OnRepeat: 1,
	}
	defaultWestRun2 = binds.KeyGroup{
		Keys:     []uint8{80},
		Pressed:  true,
		Repeat:   true,
		OnRepeat: 1,
	}
	defaultWestRunStop1 = binds.KeyGroup{
		Keys:    []uint8{104},
		Pressed: false,
	}
	defaultWestRunStop2 = binds.KeyGroup{
		Keys:    []uint8{80},
		Pressed: false,
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
		Keys:     []uint8{108},
		Pressed:  true,
		Repeat:   true,
		OnRepeat: 1,
	}
	defaultEastRun2 = binds.KeyGroup{
		Keys:     []uint8{79},
		Pressed:  true,
		Repeat:   true,
		OnRepeat: 1,
	}
	defaultEastRunStop1 = binds.KeyGroup{
		Keys:    []uint8{108},
		Pressed: false,
	}
	defaultEastRunStop2 = binds.KeyGroup{
		Keys:    []uint8{79},
		Pressed: false,
	}
	defaultUp1 = binds.KeyGroup{
		Keys:      []uint8{107},
		Modifiers: 1,
		Pressed:   true,
	}
	defaultUp2 = binds.KeyGroup{
		Keys:      []uint8{82},
		Modifiers: 1,
		Pressed:   true,
	}
	defaultUpRun1 = binds.KeyGroup{
		Keys:      []uint8{107},
		Modifiers: 1,
		Pressed:   true,
		Repeat:    true,
		OnRepeat:  1,
	}
	defaultUpRun2 = binds.KeyGroup{
		Keys:      []uint8{82},
		Modifiers: 1,
		Pressed:   true,
		Repeat:    true,
		OnRepeat:  1,
	}
	defaultUpRunStop1 = binds.KeyGroup{
		Keys:      []uint8{107},
		Modifiers: 1,
		Pressed:   false,
	}
	defaultUpRunStop2 = binds.KeyGroup{
		Keys:      []uint8{82},
		Modifiers: 1,
		Pressed:   false,
	}
	defaultDown1 = binds.KeyGroup{
		Keys:      []uint8{106},
		Modifiers: 1,
		Pressed:   true,
	}
	defaultDown2 = binds.KeyGroup{
		Keys:      []uint8{81},
		Modifiers: 1,
		Pressed:   true,
	}
	defaultDownRun1 = binds.KeyGroup{
		Keys:      []uint8{106},
		Modifiers: 1,
		Pressed:   true,
		Repeat:    true,
		OnRepeat:  1,
	}
	defaultDownRun2 = binds.KeyGroup{
		Keys:      []uint8{81},
		Modifiers: 1,
		Pressed:   true,
		Repeat:    true,
		OnRepeat:  1,
	}
	defaultDownRunStop1 = binds.KeyGroup{
		Keys:      []uint8{106},
		Modifiers: 1,
		Pressed:   false,
	}
	defaultDownRunStop2 = binds.KeyGroup{
		Keys:      []uint8{81},
		Modifiers: 1,
		Pressed:   false,
	}

	// Attack
	defaultNorthAttackRepeat1 = binds.KeyGroup{
		Keys:      []uint8{107},
		Modifiers: 1,
		Pressed:   true,
		Repeat:    true,
		OnRepeat:  1,
	}
	defaultNorthAttackRepeat2 = binds.KeyGroup{
		Keys:      []uint8{82},
		Modifiers: 1,
		Pressed:   true,
		Repeat:    true,
		OnRepeat:  1,
	}
	defaultNorthAttackStop1 = binds.KeyGroup{
		Keys:      []uint8{107},
		Modifiers: 1,
		Pressed:   false,
	}
	defaultNorthAttackStop2 = binds.KeyGroup{
		Keys:      []uint8{82},
		Modifiers: 1,
		Pressed:   false,
	}
	defaultSouthAttackRepeat1 = binds.KeyGroup{
		Keys:      []uint8{106},
		Modifiers: 1,
		Pressed:   true,
		Repeat:    true,
		OnRepeat:  1,
	}
	defaultSouthAttackRepeat2 = binds.KeyGroup{
		Keys:      []uint8{81},
		Modifiers: 1,
		Pressed:   true,
		Repeat:    true,
		OnRepeat:  1,
	}
	defaultSouthAttackStop1 = binds.KeyGroup{
		Keys:      []uint8{106},
		Modifiers: 1,
		Pressed:   false,
	}
	defaultSouthAttackStop2 = binds.KeyGroup{
		Keys:      []uint8{81},
		Modifiers: 1,
		Pressed:   false,
	}
	defaultWestAttackRepeat1 = binds.KeyGroup{
		Keys:      []uint8{104},
		Modifiers: 1,
		Pressed:   true,
		Repeat:    true,
		OnRepeat:  1,
	}
	defaultWestAttackRepeat2 = binds.KeyGroup{
		Keys:      []uint8{80},
		Modifiers: 1,
		Pressed:   true,
		Repeat:    true,
		OnRepeat:  1,
	}
	defaultWestAttackStop1 = binds.KeyGroup{
		Keys:      []uint8{104},
		Modifiers: 1,
		Pressed:   false,
	}
	defaultWestAttackStop2 = binds.KeyGroup{
		Keys:      []uint8{80},
		Modifiers: 1,
		Pressed:   false,
	}
	defaultEastAttackRepeat1 = binds.KeyGroup{
		Keys:      []uint8{108},
		Modifiers: 1,
		Pressed:   true,
		Repeat:    true,
		OnRepeat:  1,
	}
	defaultEastAttackRepeat2 = binds.KeyGroup{
		Keys:      []uint8{79},
		Modifiers: 1,
		Pressed:   true,
		Repeat:    true,
		OnRepeat:  1,
	}
	defaultEastAttackStop1 = binds.KeyGroup{
		Keys:      []uint8{108},
		Modifiers: 1,
		Pressed:   false,
	}
	defaultEastAttackStop2 = binds.KeyGroup{
		Keys:      []uint8{79},
		Modifiers: 1,
		Pressed:   false,
	}
	defaultUpAttackRepeat1 = binds.KeyGroup{
		Keys:      []uint8{107},
		Modifiers: 65,
		Pressed:   true,
		Repeat:    true,
		OnRepeat:  1,
	}
	defaultUpAttackRepeat2 = binds.KeyGroup{
		Keys:      []uint8{82},
		Modifiers: 65,
		Pressed:   true,
		Repeat:    true,
		OnRepeat:  1,
	}
	defaultUpAttackStop1 = binds.KeyGroup{
		Keys:      []uint8{107},
		Modifiers: 65,
		Pressed:   false,
	}
	defaultUpAttackStop2 = binds.KeyGroup{
		Keys:      []uint8{82},
		Modifiers: 65,
		Pressed:   false,
	}
	defaultDownAttackRepeat1 = binds.KeyGroup{
		Keys:      []uint8{106},
		Modifiers: 65,
		Pressed:   true,
		Repeat:    true,
		OnRepeat:  1,
	}
	defaultDownAttackRepeat2 = binds.KeyGroup{
		Keys:      []uint8{81},
		Modifiers: 65,
		Pressed:   true,
		Repeat:    true,
		OnRepeat:  1,
	}
	defaultDownAttackStop1 = binds.KeyGroup{
		Keys:      []uint8{106},
		Modifiers: 65,
		Pressed:   false,
	}
	defaultDownAttackStop2 = binds.KeyGroup{
		Keys:      []uint8{81},
		Modifiers: 65,
		Pressed:   false,
	}

	defaultFocusChat = binds.KeyGroup{
		Keys:    []uint8{13},
		Pressed: true,
	}
	defaultFocusCommand = binds.KeyGroup{
		Keys:    []uint8{47},
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
	// Debug
	s.bindings.SetFunction("debug", func(i ...interface{}) {
		s.DebugWindow.Toggle()
	})
	// Movement
	s.bindings.SetFunction("north", func(i ...interface{}) {
		s.runDirection = network.North
		s.Client.Send(network.CommandCmd{
			Cmd: network.North,
		})
	})
	s.bindings.SetFunction("north run", func(i ...interface{}) {
		s.runDirection = network.North
		s.Client.Send(network.CommandRepeatCmd{
			Cmd: network.North,
		})
	})
	s.bindings.SetFunction("north run stop", func(i ...interface{}) {
		s.Client.Send(network.CommandRepeatCmd{
			Cmd:    network.North,
			Cancel: true,
		})
	})
	s.bindings.SetFunction("south", func(i ...interface{}) {
		s.runDirection = network.South
		s.Client.Send(network.CommandCmd{
			Cmd: network.South,
		})
	})
	s.bindings.SetFunction("south run", func(i ...interface{}) {
		s.runDirection = network.South
		s.Client.Send(network.CommandRepeatCmd{
			Cmd: network.South,
		})
	})
	s.bindings.SetFunction("south run stop", func(i ...interface{}) {
		s.Client.Send(network.CommandRepeatCmd{
			Cmd:    network.South,
			Cancel: true,
		})
	})
	s.bindings.SetFunction("east", func(i ...interface{}) {
		s.runDirection = network.East
		s.Client.Send(network.CommandCmd{
			Cmd: network.East,
		})
	})
	s.bindings.SetFunction("east run", func(i ...interface{}) {
		s.runDirection = network.East
		s.Client.Send(network.CommandRepeatCmd{
			Cmd: network.East,
		})
	})
	s.bindings.SetFunction("east run stop", func(i ...interface{}) {
		s.Client.Send(network.CommandRepeatCmd{
			Cmd:    network.East,
			Cancel: true,
		})
	})
	s.bindings.SetFunction("west", func(i ...interface{}) {
		s.runDirection = network.West
		s.Client.Send(network.CommandCmd{
			Cmd: network.West,
		})
	})
	s.bindings.SetFunction("west run", func(i ...interface{}) {
		s.runDirection = network.West
		s.Client.Send(network.CommandRepeatCmd{
			Cmd: network.West,
		})
	})
	s.bindings.SetFunction("west run stop", func(i ...interface{}) {
		s.Client.Send(network.CommandRepeatCmd{
			Cmd:    network.West,
			Cancel: true,
		})
	})
	s.bindings.SetFunction("up", func(i ...interface{}) {
		s.runDirection = network.Up
		s.Client.Send(network.CommandCmd{
			Cmd: network.Up,
		})
	})
	s.bindings.SetFunction("up run", func(i ...interface{}) {
		s.runDirection = network.Up
		s.Client.Send(network.CommandRepeatCmd{
			Cmd: network.Up,
		})
	})
	s.bindings.SetFunction("up run stop", func(i ...interface{}) {
		s.Client.Send(network.CommandRepeatCmd{
			Cmd:    network.Up,
			Cancel: true,
		})
	})
	s.bindings.SetFunction("down", func(i ...interface{}) {
		s.runDirection = network.Down
		s.Client.Send(network.CommandCmd{
			Cmd: network.Down,
		})
	})
	s.bindings.SetFunction("down run", func(i ...interface{}) {
		s.runDirection = network.Down
		s.Client.Send(network.CommandRepeatCmd{
			Cmd: network.Down,
		})
	})
	s.bindings.SetFunction("down run stop", func(i ...interface{}) {
		s.Client.Send(network.CommandRepeatCmd{
			Cmd:    network.Down,
			Cancel: true,
		})
	})
	// Attack
	s.bindings.SetFunction("north attack", func(i ...interface{}) {
		s.Client.Send(network.CommandCmd{
			Cmd: network.Attack,
			Data: network.CommandAttack{
				Direction: network.North,
			},
		})
	})
	s.bindings.SetFunction("north attack repeat", func(i ...interface{}) {
		s.Client.Send(network.CommandRepeatCmd{
			Cmd: network.Attack,
			Data: network.CommandAttack{
				Direction: network.North,
			},
		})
	})
	s.bindings.SetFunction("north run stop", func(i ...interface{}) {
		s.Client.Send(network.CommandRepeatCmd{
			Cmd: network.Attack,
			Data: network.CommandAttack{
				Direction: network.North,
			},
			Cancel: true,
		})
	})
	s.bindings.SetFunction("south attack", func(i ...interface{}) {
		s.Client.Send(network.CommandCmd{
			Cmd: network.Attack,
			Data: network.CommandAttack{
				Direction: network.South,
			},
		})
	})
	s.bindings.SetFunction("south attack repeat", func(i ...interface{}) {
		s.Client.Send(network.CommandRepeatCmd{
			Cmd: network.Attack,
			Data: network.CommandAttack{
				Direction: network.South,
			},
		})
	})
	s.bindings.SetFunction("south attack stop", func(i ...interface{}) {
		s.Client.Send(network.CommandRepeatCmd{
			Cmd: network.Attack,
			Data: network.CommandAttack{
				Direction: network.South,
			},
			Cancel: true,
		})
	})
	s.bindings.SetFunction("east attack", func(i ...interface{}) {
		s.Client.Send(network.CommandCmd{
			Cmd: network.Attack,
			Data: network.CommandAttack{
				Direction: network.East,
			},
		})
	})
	s.bindings.SetFunction("east attack repeat", func(i ...interface{}) {
		s.runDirection = network.East
		s.Client.Send(network.CommandRepeatCmd{
			Cmd: network.Attack,
			Data: network.CommandAttack{
				Direction: network.East,
			},
		})
	})
	s.bindings.SetFunction("east attack stop", func(i ...interface{}) {
		s.Client.Send(network.CommandRepeatCmd{
			Cmd: network.Attack,
			Data: network.CommandAttack{
				Direction: network.East,
			},
			Cancel: true,
		})
	})
	s.bindings.SetFunction("west attack", func(i ...interface{}) {
		s.Client.Send(network.CommandCmd{
			Cmd: network.Attack,
			Data: network.CommandAttack{
				Direction: network.West,
			},
		})
	})
	s.bindings.SetFunction("west attack repeat", func(i ...interface{}) {
		s.Client.Send(network.CommandRepeatCmd{
			Cmd: network.Attack,
			Data: network.CommandAttack{
				Direction: network.West,
			},
		})
	})
	s.bindings.SetFunction("west attack stop", func(i ...interface{}) {
		s.Client.Send(network.CommandRepeatCmd{
			Cmd: network.Attack,
			Data: network.CommandAttack{
				Direction: network.West,
			},
			Cancel: true,
		})
	})
	s.bindings.SetFunction("up attack", func(i ...interface{}) {
		s.Client.Send(network.CommandCmd{
			Cmd: network.Attack,
			Data: network.CommandAttack{
				Direction: network.Up,
			},
		})
	})
	s.bindings.SetFunction("up attack repeat", func(i ...interface{}) {
		s.Client.Send(network.CommandRepeatCmd{
			Cmd: network.Attack,
			Data: network.CommandAttack{
				Direction: network.Up,
			},
		})
	})
	s.bindings.SetFunction("up attack stop", func(i ...interface{}) {
		s.Client.Send(network.CommandRepeatCmd{
			Cmd: network.Attack,
			Data: network.CommandAttack{
				Direction: network.Up,
			},
			Cancel: true,
		})
	})
	s.bindings.SetFunction("down attack", func(i ...interface{}) {
		s.Client.Send(network.CommandCmd{
			Cmd: network.Attack,
			Data: network.CommandAttack{
				Direction: network.Down,
			},
		})
	})
	s.bindings.SetFunction("down attack repeat", func(i ...interface{}) {
		s.Client.Send(network.CommandRepeatCmd{
			Cmd: network.Attack,
			Data: network.CommandAttack{
				Direction: network.Down,
			},
		})
	})
	s.bindings.SetFunction("down attack stop", func(i ...interface{}) {
		s.Client.Send(network.CommandRepeatCmd{
			Cmd: network.Attack,
			Data: network.CommandAttack{
				Direction: network.Down,
			},
			Cancel: true,
		})
	})

	// Other
	s.bindings.SetFunction("quit", func(i ...interface{}) {
		os.Exit(0)
	})
	s.bindings.SetFunction("clear commands", func(i ...interface{}) {
		s.Client.Send(network.CommandClearCmd{})
	})
	s.bindings.SetFunction("disconnect", func(i ...interface{}) {
		go func() {
			s.inputChan <- DisconnectEvent{}
		}()
	})
	s.bindings.SetFunction("wizard", func(i ...interface{}) {
		s.Client.Send(network.CommandCmd{
			Cmd: network.Wizard,
		})
	})
	s.bindings.SetFunction("wiz", func(i ...interface{}) {
		var str string
		if len(i) > 0 {
			for _, v := range i {
				switch v := v.(type) {
				case string:
					str = v
				case []string:
					str = strings.Join(v, " ")
				}
			}
		}
		s.Client.Send(network.CommandExtCmd{
			Cmd:  "wiz",
			Args: strings.Split(str, " "),
		})
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

	s.bindings.SetFunction("clear focus", func(i ...interface{}) {
		s.FocusObject(0)
	})

	s.bindings.SetFunction("focus chat", func(i ...interface{}) {
		s.ChatInput.GetUpdateChannel() <- ui.UpdateFocus{}
	})
	s.bindings.SetFunction("focus cmd", func(i ...interface{}) {
		s.ChatInput.GetUpdateChannel() <- ui.UpdateFocus{}
		s.ChatInput.GetUpdateChannel() <- ui.UpdateValue{Value: "/"}
	})

	if len(s.bindings.Keygroups) == 0 {
		if !s.bindings.HasKeygroupsForName("clear commands") {
			s.bindings.AddKeygroup("clear commands", defaultClearCommands)
		}
		if !s.bindings.HasKeygroupsForName("north") {
			s.bindings.AddKeygroup("north", defaultNorth1)
			s.bindings.AddKeygroup("north", defaultNorth2)
		}
		if !s.bindings.HasKeygroupsForName("north run") {
			s.bindings.AddKeygroup("north run", defaultNorthRun1)
			s.bindings.AddKeygroup("north run", defaultNorthRun2)
		}
		if !s.bindings.HasKeygroupsForName("north run stop") {
			s.bindings.AddKeygroup("north run stop", defaultNorthRunStop1)
			s.bindings.AddKeygroup("north run stop", defaultNorthRunStop2)
		}
		if !s.bindings.HasKeygroupsForName("south") {
			s.bindings.AddKeygroup("south", defaultSouth1)
			s.bindings.AddKeygroup("south", defaultSouth2)
		}
		if !s.bindings.HasKeygroupsForName("south run") {
			s.bindings.AddKeygroup("south run", defaultSouthRun1)
			s.bindings.AddKeygroup("south run", defaultSouthRun2)
		}
		if !s.bindings.HasKeygroupsForName("south run stop") {
			s.bindings.AddKeygroup("south run stop", defaultSouthRunStop1)
			s.bindings.AddKeygroup("south run stop", defaultSouthRunStop2)
		}
		if !s.bindings.HasKeygroupsForName("west") {
			s.bindings.AddKeygroup("west", defaultWest1)
			s.bindings.AddKeygroup("west", defaultWest2)
		}
		if !s.bindings.HasKeygroupsForName("west run") {
			s.bindings.AddKeygroup("west run", defaultWestRun1)
			s.bindings.AddKeygroup("west run", defaultWestRun2)
		}
		if !s.bindings.HasKeygroupsForName("west run stop") {
			s.bindings.AddKeygroup("west run stop", defaultWestRunStop1)
			s.bindings.AddKeygroup("west run stop", defaultWestRunStop2)
		}
		if !s.bindings.HasKeygroupsForName("east") {
			s.bindings.AddKeygroup("east", defaultEast1)
			s.bindings.AddKeygroup("east", defaultEast2)
		}
		if !s.bindings.HasKeygroupsForName("east run") {
			s.bindings.AddKeygroup("east run", defaultEastRun1)
			s.bindings.AddKeygroup("east run", defaultEastRun2)
		}
		if !s.bindings.HasKeygroupsForName("east run stop") {
			s.bindings.AddKeygroup("east run stop", defaultEastRunStop1)
			s.bindings.AddKeygroup("east run stop", defaultEastRunStop2)
		}
		if !s.bindings.HasKeygroupsForName("up") {
			s.bindings.AddKeygroup("up", defaultUp1)
			s.bindings.AddKeygroup("up", defaultUp2)
		}
		if !s.bindings.HasKeygroupsForName("up run") {
			s.bindings.AddKeygroup("up run", defaultUpRun1)
			s.bindings.AddKeygroup("up run", defaultUpRun2)
		}
		if !s.bindings.HasKeygroupsForName("up run stop") {
			s.bindings.AddKeygroup("up run stop", defaultUpRunStop1)
			s.bindings.AddKeygroup("up run stop", defaultUpRunStop2)
		}
		if !s.bindings.HasKeygroupsForName("down") {
			s.bindings.AddKeygroup("down", defaultDown1)
			s.bindings.AddKeygroup("down", defaultDown2)
		}
		if !s.bindings.HasKeygroupsForName("down run") {
			s.bindings.AddKeygroup("down run", defaultDownRun1)
			s.bindings.AddKeygroup("down run", defaultDownRun2)
		}
		if !s.bindings.HasKeygroupsForName("down run stop") {
			s.bindings.AddKeygroup("down run stop", defaultDownRunStop1)
			s.bindings.AddKeygroup("down run stop", defaultDownRunStop2)
		}

		if !s.bindings.HasKeygroupsForName("north attack repeat") {
			s.bindings.AddKeygroup("north attack repeat", defaultNorthAttackRepeat1)
			s.bindings.AddKeygroup("north attack repeat", defaultNorthAttackRepeat2)
		}
		if !s.bindings.HasKeygroupsForName("north attack stop") {
			s.bindings.AddKeygroup("north attack stop", defaultNorthAttackStop1)
			s.bindings.AddKeygroup("north attack stop", defaultNorthAttackStop2)
		}
		if !s.bindings.HasKeygroupsForName("south attack repeat") {
			s.bindings.AddKeygroup("south attack repeat", defaultSouthAttackRepeat1)
			s.bindings.AddKeygroup("south attack repeat", defaultSouthAttackRepeat2)
		}
		if !s.bindings.HasKeygroupsForName("south attack stop") {
			s.bindings.AddKeygroup("south attack stop", defaultSouthAttackStop1)
			s.bindings.AddKeygroup("south attack stop", defaultSouthAttackStop2)
		}
		if !s.bindings.HasKeygroupsForName("west attack repeat") {
			s.bindings.AddKeygroup("west attack repeat", defaultWestAttackRepeat1)
			s.bindings.AddKeygroup("west attack repeat", defaultWestAttackRepeat2)
		}
		if !s.bindings.HasKeygroupsForName("west attack stop") {
			s.bindings.AddKeygroup("west attack stop", defaultWestAttackStop1)
			s.bindings.AddKeygroup("west attack stop", defaultWestAttackStop2)
		}
		if !s.bindings.HasKeygroupsForName("east attack repeat") {
			s.bindings.AddKeygroup("east attack repeat", defaultEastAttackRepeat1)
			s.bindings.AddKeygroup("east attack repeat", defaultEastAttackRepeat2)
		}
		if !s.bindings.HasKeygroupsForName("east attack stop") {
			s.bindings.AddKeygroup("east attack stop", defaultEastAttackStop1)
			s.bindings.AddKeygroup("east attack stop", defaultEastAttackStop2)
		}
		if !s.bindings.HasKeygroupsForName("up attack repeat") {
			s.bindings.AddKeygroup("up attack repeat", defaultUpAttackRepeat1)
			s.bindings.AddKeygroup("up attack repeat", defaultUpAttackRepeat2)
		}
		if !s.bindings.HasKeygroupsForName("up attack stop") {
			s.bindings.AddKeygroup("up attack stop", defaultUpAttackStop1)
			s.bindings.AddKeygroup("up attack stop", defaultUpAttackStop2)
		}
		if !s.bindings.HasKeygroupsForName("down attack repeat") {
			s.bindings.AddKeygroup("down attack repeat", defaultDownAttackRepeat1)
			s.bindings.AddKeygroup("down attack repeat", defaultDownAttackRepeat2)
		}
		if !s.bindings.HasKeygroupsForName("down attack stop") {
			s.bindings.AddKeygroup("down attack stop", defaultDownAttackStop1)
			s.bindings.AddKeygroup("down attack stop", defaultDownAttackStop2)
		}

		if !s.bindings.HasKeygroupsForName("clear focus") {
			s.bindings.AddKeygroup("clear focus", defaultClearFocus)
		}

		if !s.bindings.HasKeygroupsForName("focus chat") {
			s.bindings.AddKeygroup("focus chat", defaultFocusChat)
		}
		if !s.bindings.HasKeygroupsForName("focus cmd") {
			s.bindings.AddKeygroup("focus cmd", defaultFocusCommand)
		}
	}
}
