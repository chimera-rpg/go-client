package states

import (
	"time"

	"github.com/chimera-rpg/go-client/binds"
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/data"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-client/world"
	"github.com/chimera-rpg/go-common/network"
)

type CommandMode = int

const (
	CommandModeChat = iota
	CommandModeSay
	CommandModeCmd
)

var CommandModeStrings = []string{
	"CHAT",
	"SAY",
	"CMD",
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
	world            world.World
	keyBinds         []uint8
	inputChan        chan UserInput // This channel is used to transfer input from the UI goroutine to the Client goroutine safely.
	objectImages     map[uint32]ui.ElementI
	objectImageIDs   map[uint32]data.StringID
	mapMessages      []MapMessage
	bindings         *binds.Bindings
}

// Init our Game state.
func (s *Game) Init(t interface{}) (state client.StateI, nextArgs interface{}, err error) {
	s.inputChan = make(chan UserInput)
	s.objectImages = make(map[uint32]ui.ElementI)
	s.objectImageIDs = make(map[uint32]data.StringID)
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
	s.CleanupUI()
}

// Loop is our loop for managing network activity and beyond.
func (s *Game) Loop() {
	cleanupChan := make(chan struct{})
	go func() {
		for {
			time.Sleep(time.Second * 1)
			cleanupChan <- struct{}{}
		}
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
					s.Client.Send(network.CommandMessage{
						Type: network.ChatMessage,
						Body: e.Body,
					})
				}
			case MouseInput:
				s.Client.Log.Printf("mouse: %+v\n", e)
			case ChangeCommandMode:
				s.CommandMode++
				if s.CommandMode >= len(CommandModeStrings) {
					s.CommandMode = 0
				}
				s.ChatType.GetUpdateChannel() <- ui.UpdateValue{Value: CommandModeStrings[s.CommandMode]}
			case DisconnectEvent:
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
		s.Client.HandleMessageCommand(c)
		s.UpdateMessagesWindow()
	default:
		s.Client.Log.Printf("Server sent a Command %+v\n", c)
	}
	return false
}
