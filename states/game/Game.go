package game

import (
	"fmt"
	"image/color"
	"time"

	"github.com/chimera-rpg/go-client/audio"
	"github.com/chimera-rpg/go-client/binds"
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/config"
	"github.com/chimera-rpg/go-client/states/game/elements"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-client/world"
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-common/network"
	"github.com/veandco/go-sdl2/sdl"
)

type CommandMode = int

const (
	CommandModeChat = iota
	CommandModeSay

// CommandModeCmd
)

var CommandModeStrings = []string{
	"CHAT",
	"SAY",
	//	"CMD",
}

// Game is our live Game state, used once the user has connected to the server
// and joined as a player character.
type Game struct {
	client.State
	CommandMode          CommandMode
	GameContainer        ui.Container
	MessagesWindow       ui.Container
	ChatType             ui.ElementI
	ChatInput            ui.ElementI
	ChatWindow           ui.Container
	messageElements      []ui.ElementI
	CommandContainer     ui.ElementI
	InventoryWindow      ui.Container
	InspectorWindow      elements.InspectorWindow
	MapWindow            elements.MapWindow
	GroundWindow         elements.GroundModeWindow
	StatsWindow          ui.Container
	StateWindow          ui.Container
	statusElements       map[cdata.StatusType]ui.ElementI
	statuses             map[cdata.StatusType]bool
	world                world.World
	keyBinds             []uint8
	inputChan            chan interface{} // This channel is used to transfer input from the UI goroutine to the Client goroutine safely.
	objectShadows        map[uint32]ui.ElementI
	MessageHistory       []Message
	bindings             *binds.Bindings
	repeatingKeys        map[uint8]int
	heldButtons          map[uint8]bool
	runDirection         int
	objectsScale         *float64               // Pointer to config graphics.
	pendingNoiseCommands []network.CommandNoise // Pending noises, for sounds that have not loaded yet.
	pendingMusicCommands []network.CommandMusic // Pending music, for sounds that have not loaded yet.
	focusedObjectID      uint32
	hoveredObjectID      uint32
	focusedImage         ui.ElementI
	eventHooks           map[interface{}][]func(e interface{})
}

// Init our Game state.
func (s *Game) Init(t interface{}) (state client.StateI, nextArgs interface{}, err error) {
	s.inputChan = make(chan interface{})
	s.objectShadows = make(map[uint32]ui.ElementI)
	s.statuses = make(map[cdata.StatusType]bool)
	s.statusElements = make(map[cdata.StatusType]ui.ElementI)
	s.repeatingKeys = make(map[uint8]int)
	s.heldButtons = make(map[uint8]bool)
	s.SetupBinds()
	s.CommandMode = CommandModeChat
	// Initialize our world.
	s.world.Init(s.Client.DataManager, s.Client.Log)
	s.eventHooks = make(map[interface{}][]func(e interface{}))

	// This is lazy(tm), but we're just resending all pendingNoiseCommands on receipt of a sound or audio network command.
	s.Client.DataManager.SetHandleCallback(func(netID int, cmd network.Command) {
		if netID == network.TypeSound || netID == network.TypeAudio {
			if len(s.pendingNoiseCommands) > 0 {
				pending := s.pendingNoiseCommands
				s.pendingNoiseCommands = make([]network.CommandNoise, 0)
				for _, c := range pending {
					s.HandleNet(c)
				}
			}
			if len(s.pendingMusicCommands) > 0 {
				pending := s.pendingMusicCommands
				s.pendingMusicCommands = make([]network.CommandMusic, 0)
				for _, c := range pending {
					s.HandleNet(c)
				}
			}
		} else if netID == network.TypeAnimation {
			c := cmd.(network.CommandAnimation)
			s.world.CheckPendingObjectAnimations(c.AnimationID)
		} else if netID == network.TypeGraphics {
			c := cmd.(network.CommandGraphics)
			s.world.CheckPendingObjectImageIDs(c.GraphicsID)
		}
	})

	s.Client.Log.Print("Game State")

	s.SetupUI()

	go s.Loop()
	return
}

// Close our Game state.
func (s *Game) Close() {
	go func() {
		s.Client.Connection.Close()
	}()
	s.CleanupUI()
	s.Client.Audio.CommandChannel <- audio.CommandStopAllMusic{}
}

// Loop is our loop for managing network activity and beyond.
func (s *Game) Loop() {
	ticker := time.NewTicker(16 * time.Millisecond)
	lastTs := time.Now()
	for {
		ts := time.Now()
		delta := ts.Sub(lastTs)
		select {
		case cmd := <-s.Client.CmdChan:
			ret := s.HandleNet(cmd)
			if ret {
				return
			}
		case <-s.Client.ClosedChan:
			s.Client.Log.Print("Lost connection to server.")
			ticker.Stop()
			s.Client.StateChannel <- client.StateMessage{PopToTop: true, Args: nil}
			return
		case inp := <-s.inputChan:
			switch e := inp.(type) {
			case elements.ResizeEvent:
				s.UpdateMessagesWindow()
			case KeyInput:
				if !e.pressed {
					s.repeatingKeys[e.code] = 0
				}
				// Remove
				s.bindings.Trigger(binds.KeyGroup{
					Keys:      []uint8{e.code},
					Modifiers: e.modifiers &^ sdl.KMOD_NUM, // Remove numlock as a modifier
					Pressed:   e.pressed,
					Repeat:    e.repeat,
					OnRepeat:  s.repeatingKeys[e.code],
				}, nil)
				if e.pressed && e.repeat {
					s.repeatingKeys[e.code]++
				}
			case elements.ChatEvent:
				if s.isChatCommand(e.Body) {
					s.processChatCommand(e.Body)
				} else {
					if s.CommandMode == CommandModeChat {
						s.Client.Send(network.CommandMessage{
							Type: network.ChatMessage,
							Body: e.Body,
						})

					} else if s.CommandMode == CommandModeSay {
						s.Client.Send(network.CommandMessage{
							Type: network.PCMessage,
							Body: e.Body,
						})
					}
				}
			case elements.GroundModeChangeEvent:
				for _, cb := range s.eventHooks[elements.GroundModeChangeEvent{}] {
					cb(e)
				}
			case elements.GroundModeComboEvent:
				for _, cb := range s.eventHooks[elements.GroundModeComboEvent{}] {
					cb(e)
				}
			case elements.MouseInput:
				if e.Button == 3 {
					s.MoveWithMouse(e)
				}
			case elements.MouseMoveInput:
				if s.heldButtons[3] {
					s.RunWithMouse(e.X, e.Y)
				}
			case elements.FocusObjectEvent:
				s.FocusObject(e.ID)
				for _, cb := range s.eventHooks[elements.FocusObjectEvent{}] {
					cb(e)
				}
			case elements.HoverObjectEvent:
				s.HoverObject(e.ID)
				for _, cb := range s.eventHooks[elements.HoverObjectEvent{}] {
					cb(e)
				}
			case elements.UnhoverObjectEvent:
				s.UnhoverObject(e.ID)
			case ChangeCommandMode:
				s.CommandMode++
				if s.CommandMode >= len(CommandModeStrings) {
					s.CommandMode = 0
				}
				s.ChatType.GetUpdateChannel() <- ui.UpdateValue{Value: CommandModeStrings[s.CommandMode]}
			case DisconnectEvent:
				s.Client.Log.Print("Disconnected from server.")
				ticker.Stop()
				s.Client.StateChannel <- client.StateMessage{PopToTop: true, Args: nil}
				return
			}
		case <-ticker.C:
		}
		s.HandleRender(delta)
		s.UpdateGroundWindow()
		lastTs = ts
	}
}

// HandleNet handles the network code for our Game state.
func (s *Game) HandleNet(cmd network.Command) bool {
	switch c := cmd.(type) {
	case network.CommandGraphics:
		s.Client.DataManager.HandleGraphicsCommand(c)
	case network.CommandAnimation:
		s.Client.DataManager.HandleAnimationCommand(c)
	case network.CommandSound:
		s.Client.DataManager.HandleSoundCommand(c)
	case network.CommandAudio:
		s.Client.DataManager.HandleAudioCommand(c)
	case network.CommandMap:
		// FIXME: We probably should handle removing of elements this elsewhere.
		{
			var batchMessages BatchMessages

			for _, o := range s.world.GetObjects() {
				if o == s.world.GetViewObject() {
					continue
				}
				if o.Element != nil {
					batchMessages.add(ui.BatchDestroyMessage{
						Target: o.Element,
					})
				}
				if o.ShadowElement != nil {
					batchMessages.add(ui.BatchDestroyMessage{
						Target: o.ShadowElement,
					})
				}
			}
			s.Client.RootWindow.BatchChannel <- batchMessages.messages
		}
		s.world.HandleMapCommand(c)
	case network.CommandObject:
		s.world.HandleObjectCommand(c)
	case network.CommandTile:
		s.world.HandleTileCommand(c)
	case network.CommandTileLight:
		s.world.HandleTileLightCommand(c)
	case network.CommandMessage:
		s.HandleMessageCommand(c)
		s.UpdateMessagesWindow()
	case network.CommandStatus:
		// FIXME: Move
		if c.Type == cdata.SqueezingStatus {
			s.world.GetViewObject().Squeezing = c.Active
			s.world.GetViewObject().Changed = true
		} else if c.Type == cdata.CrouchingStatus {
			s.world.GetViewObject().Crouching = c.Active
			s.world.GetViewObject().Changed = true
		}
		s.statuses[c.Type] = c.Active
		s.UpdateStateWindow()
	case network.CommandNoise:
		s.world.HandleNoiseCommand(c)
		snd, ok := s.Client.DataManager.GetAudioSound(c.AudioID, c.SoundID, 0)
		if !ok {
			s.pendingNoiseCommands = append(s.pendingNoiseCommands, c)
		} else {
			s.Client.Audio.CommandChannel <- audio.CommandPlaySound{
				ID:     snd.SoundID,
				Volume: c.Volume,
			}
			if m, err := s.createMapMessage(int(c.Y), int(c.X), int(c.Z), "*"+snd.Text+"*", color.RGBA{128, 200, 255, 220}); err == nil {
				s.MapWindow.Messages = append(s.MapWindow.Messages, m)
				s.MapWindow.Container.GetAdoptChannel() <- m.El
			}
		}
	case network.CommandMusic:
		s.Client.DataManager.EnsureAudio(c.AudioID)
		snd, ok := s.Client.DataManager.GetAudioSound(c.AudioID, c.SoundID, 0)
		if !ok {
			s.pendingMusicCommands = append(s.pendingMusicCommands, c)
		} else {
			s.Client.Audio.CommandChannel <- audio.CommandPlayMusic{
				ID:         snd.SoundID,
				PlaybackID: c.ObjectID,
				Volume:     c.Volume,
				//Loop:       c.Loop,
			}
			// TODO: Some sort of "you hear music..." then add some credits?
			/*if m, err := s.createMapMessage(c.Y, c.X, c.Z, "*"+snd.Text+"*", color.RGBA{128, 200, 255, 220}); err == nil {
				s.mapMessages = append(s.mapMessages, m)
				s.MapContainer.GetAdoptChannel() <- m.el
			}*/
		}
	case network.CommandDamage:
		// TODO: Limit damage indicators to _only_ within visible range!
		var totalDamage float64
		for _, d := range c.StyleDamage {
			totalDamage += d
		}
		totalDamage += c.AttributeDamage
		if m, err := s.createMapObjectMessage(c.Target, fmt.Sprintf("%1.f", totalDamage), color.RGBA{255, 255, 255, 200}); err == nil {
			m.FloatY = -0.02
			// TODO: Make some sort of color map for: damage types, as well as if we're the target of the damage.
			if c.Target == s.world.GetViewObject().ID {
				m.El.GetStyle().ForegroundColor = color.NRGBA{
					R: 255,
					G: 64,
					B: 64,
					A: 200,
				}
				m.El.GetStyle().OutlineColor = color.NRGBA{
					R: 255,
					G: 255,
					B: 255,
					A: 200,
				}
			} else {
				m.El.GetStyle().OutlineColor = color.NRGBA{
					R: 255,
					G: 64,
					B: 64,
					A: 200,
				}
			}
			m.El.GetStyle().ZIndex.Value = float64(99999 + len(s.MapWindow.Messages) + 1)
			s.MapWindow.Messages = append(s.MapWindow.Messages, m)
			s.MapWindow.Container.GetAdoptChannel() <- m.El
		}
		if c.Target == s.world.GetViewObject().ID {
			// TODO: Show info about us getting hit.
		} else {
			// TODO: Print damage value in combat log or some such.
		}
	default:
		s.Client.Log.Printf("Server sent a Command %+v\n", c)
	}
	// Eh... update outline on any of these changes.
	//s.focusObject(s.focusedObjectID)
	return false
}

// HandleMessageCommand received network.CommandMessage types and adds it to the client's message history.
func (s *Game) HandleMessageCommand(m network.CommandMessage) {
	s.MessageHistory = append(s.MessageHistory, Message{
		Received: time.Now(),
		Message:  m,
	})
	s.UpdateMessagesWindow()
}

func (s *Game) RunWithMouse(x, y int32) {
	dA := s.MapWindow.MouseAngleFromView(x, y)
	if dA >= 315 || dA <= 45 {
		if s.runDirection != network.East {
			s.bindings.RunFunction("east run")
		}
	} else if dA > 45 && dA <= 135 {
		if s.runDirection != network.South {
			s.bindings.RunFunction("south run")
		}
	} else if dA > 135 && dA <= 225 {
		if s.runDirection != network.West {
			s.bindings.RunFunction("west run")
		}
	} else if dA > 225 && dA <= 315 {
		if s.runDirection != network.North {
			s.bindings.RunFunction("north run")
		}
	}
}

func (s *Game) MoveWithMouse(e elements.MouseInput) {
	dA := s.MapWindow.MouseAngleFromView(e.X, e.Y)
	/****
	    	275
	  225 		315
	180  	 o	 360
	  135 	 	45
		 		90
	******/
	if e.Held {
		s.heldButtons[3] = true
		if dA >= 315 || dA <= 45 {
			s.bindings.RunFunction("east run")
		} else if dA > 45 && dA <= 135 {
			s.bindings.RunFunction("south run")
		} else if dA > 135 && dA <= 225 {
			s.bindings.RunFunction("west run")
		} else if dA > 225 && dA <= 315 {
			s.bindings.RunFunction("north run")
		}
	} else if e.Released {
		s.heldButtons[3] = false
		if dA >= 315 || dA <= 45 {
			s.bindings.RunFunction("east run stop")
		} else if dA > 45 && dA <= 135 {
			s.bindings.RunFunction("south run stop")
		} else if dA > 135 && dA <= 225 {
			s.bindings.RunFunction("west run stop")
		} else if dA > 225 && dA <= 315 {
			s.bindings.RunFunction("north run stop")
		}
	} else if !e.Pressed {
		if dA >= 315 || dA <= 45 {
			s.bindings.RunFunction("east")
		} else if dA > 45 && dA <= 135 {
			s.bindings.RunFunction("south")
		} else if dA > 135 && dA <= 225 {
			s.bindings.RunFunction("west")
		} else if dA > 225 && dA <= 315 {
			s.bindings.RunFunction("north")
		}
	}
}

func (s *Game) getObjectShadow(id uint32) ui.ElementI {
	return s.objectShadows[id]
}

func (s *Game) World() *world.World {
	return &s.world
}

func (s *Game) FocusObject(e uint32) {
	if s.focusedObjectID == s.hoveredObjectID && e != s.hoveredObjectID {
		s.focusedObjectID = e
		s.HoverObject(s.focusedObjectID)
	} else if o := s.world.GetObject(s.focusedObjectID); o != nil {
		if o.Element != nil {
			o.Element.GetUpdateChannel() <- ui.UpdateOutlineColor{0, 0, 0, 0}
		}
	}
	if o := s.world.GetObject(e); o != nil {
		if o.Element != nil {
			o.Element.GetUpdateChannel() <- ui.UpdateOutlineColor{255, 255, 0, 128}
		}
	}
	s.focusedObjectID = e
}

func (s *Game) FocusedImage() ui.ElementI {
	return s.focusedImage
}

func (s *Game) FocusedObject() *world.Object {
	return s.world.GetObject(s.focusedObjectID)
}

func (s *Game) FocusedObjectID() uint32 {
	return s.focusedObjectID
}

func (s *Game) HoverObject(e uint32) {
	if e == s.focusedObjectID || e == s.hoveredObjectID {
		return
	}
	if s.hoveredObjectID != s.focusedObjectID {
		if o := s.world.GetObject(s.hoveredObjectID); o != nil {
			if o.Element != nil {
				o.Element.GetUpdateChannel() <- ui.UpdateOutlineColor{0, 0, 0, 0}
			}
		}
	}
	if o := s.world.GetObject(e); o != nil {
		if o.Element != nil {
			o.Element.GetUpdateChannel() <- ui.UpdateOutlineColor{200, 200, 200, 128}
		}
	}

	s.hoveredObjectID = e
}

func (s *Game) UnhoverObject(e uint32) {
	if e == s.focusedObjectID || e != s.hoveredObjectID {
		return
	}
	if o := s.world.GetObject(s.hoveredObjectID); o != nil {
		if o.Element != nil {
			o.Element.GetUpdateChannel() <- ui.UpdateOutlineColor{0, 0, 0, 0}
		}
	}
	s.hoveredObjectID = 0
}

func (s *Game) HoveredObject() *world.Object {
	return s.world.GetObject(s.hoveredObjectID)
}

func (s *Game) HoveredObjectID() uint32 {
	return s.hoveredObjectID
}

func (s *Game) HookEvent(k interface{}, cb func(e interface{})) {
	if _, ok := s.eventHooks[k]; !ok {
		s.eventHooks[k] = make([]func(e interface{}), 0)
	}
	s.eventHooks[k] = append(s.eventHooks[k], cb)
}

func (s *Game) InputChan() chan interface{} {
	return s.inputChan
}

func (s *Game) SendNetMessage(cmd network.Command) {
	s.Client.Send(cmd)
}

func (s *Game) Config() *config.Config {
	return &s.Client.DataManager.Config
}
