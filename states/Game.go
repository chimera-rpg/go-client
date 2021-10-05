package states

import (
	"math"
	"time"

	"github.com/chimera-rpg/go-client/binds"
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/data"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-client/world"
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-common/network"
)

type CommandMode = int

const (
	CommandModeChat = iota
	CommandModeSay

//	CommandModeCmd
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
	CommandMode      CommandMode
	GameContainer    ui.Container
	MessagesWindow   ui.Container
	ChatType         ui.ElementI
	ChatInput        ui.ElementI
	ChatWindow       ui.Container
	messageElements  []ui.ElementI
	CommandContainer ui.ElementI
	MapContainer     ui.Container
	InventoryWindow  ui.Container
	GroundWindow     ui.Container
	StatsWindow      ui.Container
	StateWindow      ui.Container
	statusElements   map[cdata.StatusType]ui.ElementI
	statuses         map[cdata.StatusType]bool
	world            world.World
	keyBinds         []uint8
	inputChan        chan UserInput // This channel is used to transfer input from the UI goroutine to the Client goroutine safely.
	objectImages     map[uint32]ui.ElementI
	objectImageIDs   map[uint32]data.StringID
	mapMessages      []MapMessage
	MessageHistory   []Message
	bindings         *binds.Bindings
}

// Init our Game state.
func (s *Game) Init(t interface{}) (state client.StateI, nextArgs interface{}, err error) {
	s.inputChan = make(chan UserInput)
	s.objectImages = make(map[uint32]ui.ElementI)
	s.objectImageIDs = make(map[uint32]data.StringID)
	s.statuses = make(map[cdata.StatusType]bool)
	s.statusElements = make(map[cdata.StatusType]ui.ElementI)
	s.SetupBinds()
	s.CommandMode = CommandModeChat
	// Initialize our world.
	s.world.Init(s.Client.DataManager, s.Client.Log)

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
}

// Loop is our loop for managing network activity and beyond.
func (s *Game) Loop() {
	cleanupChan := make(chan struct{})
	cleanupChanQuit := make(chan struct{})
	go func() {
		for {
			select {
			case <-cleanupChanQuit:
				return
			default:
				cleanupChan <- struct{}{}
				time.Sleep(time.Second * 1)
			}
		}
	}()
	defer func() {
		cleanupChanQuit <- struct{}{}
	}()
	for {
		select {
		case cmd := <-s.Client.CmdChan:
			ret := s.HandleNet(cmd)
			if ret {
				return
			}
		case <-s.Client.ClosedChan:
			s.Client.Log.Print("Lost connection to server.")
			s.Client.StateChannel <- client.StateMessage{State: &List{}, Args: nil}
			return
		case inp := <-s.inputChan:
			switch e := inp.(type) {
			case ResizeEvent:
				s.UpdateMessagesWindow()
			case KeyInput:
				s.bindings.Trigger(binds.KeyGroup{
					Keys:      []uint8{e.code},
					Modifiers: e.modifiers,
					Pressed:   e.pressed,
					Repeat:    e.repeat,
				}, nil)
			case ChatEvent:
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
			case MouseInput:
				if e.button == 3 {
					s.MoveWithMouse(e)
				}
			case ChangeCommandMode:
				s.CommandMode++
				if s.CommandMode >= len(CommandModeStrings) {
					s.CommandMode = 0
				}
				s.ChatType.GetUpdateChannel() <- ui.UpdateValue{Value: CommandModeStrings[s.CommandMode]}
			case DisconnectEvent:
				s.Client.Log.Print("Disconnected from server.")
				s.Client.StateChannel <- client.StateMessage{State: &List{}, Args: nil}
				return
			}
		case <-cleanupChan:
		}
		s.HandleRender()
	}
}

// HandleNet handles the network code for our Game state.
func (s *Game) HandleNet(cmd network.Command) bool {
	switch c := cmd.(type) {
	case network.CommandGraphics:
		s.Client.DataManager.HandleGraphicsCommand(c)
	case network.CommandAnimation:
		s.Client.DataManager.HandleAnimationCommand(c)
	case network.CommandMap:
		s.world.HandleMapCommand(c)
	case network.CommandObject:
		s.world.HandleObjectCommand(c)
	case network.CommandTile:
		s.world.HandleTileCommand(c)
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
	default:
		s.Client.Log.Printf("Server sent a Command %+v\n", c)
	}
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

func (s *Game) MoveWithMouse(e MouseInput) {
	x1 := e.x - s.MapContainer.GetAbsoluteX()
	y1 := e.y - s.MapContainer.GetAbsoluteY()
	x2 := s.MapContainer.GetWidth() / 2
	y2 := s.MapContainer.GetHeight() / 2
	dY := y2 - y1
	dX := x2 - x1
	dA := (math.Atan2(float64(dY), float64(dX)) * 180 / math.Pi) + 180
	/****
	    	275
	  225 		315
	180  	 o	 360
	  135 	 	45
		 		90
	******/
	if !e.pressed {
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
