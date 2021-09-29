package states

import (
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/data"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-client/world"
	"github.com/chimera-rpg/go-common/network"
)

// Game is our live Game state, used once the user has connected to the server
// and joined as a player character.
type Game struct {
	client.State
	GameContainer   ui.Container
	MessagesWindow  ui.Container
	ChatInput       ui.ElementI
	ChatWindow      ui.Container
	messageElements []ui.ElementI
	MapContainer    ui.Container
	InventoryWindow ui.Container
	GroundWindow    ui.Container
	StatsWindow     ui.Container
	StateWindow     ui.Container
	world           world.World
	keyBinds        []uint8
	inputChan       chan UserInput // This channel is used to transfer input from the UI goroutine to the Client goroutine safely.
	objectImages    map[uint32]ui.ElementI
	objectImageIDs  map[uint32]data.StringID
}

// Init our Game state.
func (s *Game) Init(t interface{}) (state client.StateI, nextArgs interface{}, err error) {
	s.inputChan = make(chan UserInput)
	s.objectImages = make(map[uint32]ui.ElementI)
	s.objectImageIDs = make(map[uint32]data.StringID)
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
				// TODO: Move to key bind system.
				if e.pressed && !e.repeat {
					if e.code == 107 || e.code == 82 { // up
						s.Client.Log.Println("send north")
						s.Client.Send(network.CommandCmd{
							Cmd: network.North,
						})
					} else if e.code == 106 || e.code == 81 { // down
						s.Client.Log.Println("send south")
						s.Client.Send(network.CommandCmd{
							Cmd: network.South,
						})
					} else if e.code == 104 || e.code == 80 { // left
						s.Client.Log.Println("send west")
						s.Client.Send(network.CommandCmd{
							Cmd: network.West,
						})
					} else if e.code == 108 || e.code == 79 { // right
						s.Client.Log.Println("send east")
						s.Client.Send(network.CommandCmd{
							Cmd: network.East,
						})
					} else if e.code == 13 { // enter
						s.ChatInput.GetUpdateChannel() <- ui.UpdateFocus{}
					}
				}
			case ChatEvent:
				s.Client.Send(network.CommandMessage{
					Type: network.ChatMessage,
					Body: e.Body,
				})
			case MouseInput:
				s.Client.Log.Printf("mouse: %+v\n", e)
			}
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
