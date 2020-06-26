package states

import (
	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-client/world"
	"github.com/chimera-rpg/go-common/network"
)

// Game is our live Game state, used once the user has connected to the server
// and joined as a player character.
type Game struct {
	client.State
	GameContainer   ui.Container
	ChatWindow      ui.Container
	MapElement      ui.MapElement
	InventoryWindow ui.Container
	GroundWindow    ui.Container
	StatsWindow     ui.Container
	StateWindow     ui.Container
	world           world.World
	keyBinds        []uint8
	inputChan       chan UserInput // This channel is used to transfer input from the UI goroutine to the Client goroutine safely.
	objectImages    map[uint32]ui.ImageElement
}

// UserInput is an interface used in a channel in Game for handling UI input.
type UserInput interface {
}

// KeyInput is the Userinput for key events.
type KeyInput struct {
	code      uint8
	modifiers uint16
	pressed   bool
	repeat    bool
}

// MouseInput is the UserInput for mouse events.
type MouseInput struct {
	x, y    int32
	button  uint8
	pressed bool
}

// Init our Game state.
func (s *Game) Init(t interface{}) (state client.StateI, nextArgs interface{}, err error) {
	s.inputChan = make(chan UserInput)
	// Initialize our world.
	s.world.Init(s.Client.DataManager, s.Client.Log)

	s.Client.Log.Print("Game State")

	// Main Container
	err = s.GameContainer.Setup(ui.ContainerConfig{
		Value: "Game",
		Style: `
			W 100%
			H 100%
			BackgroundColor 139 186 139 255
		`,
		Events: ui.Events{
			OnKeyDown: func(char uint8, modifiers uint16, repeat bool) bool {
				s.inputChan <- KeyInput{
					code:      char,
					modifiers: modifiers,
					pressed:   true,
					repeat:    repeat,
				}
				return true
			},
			OnKeyUp: func(char uint8, modifiers uint16) bool {
				s.inputChan <- KeyInput{
					code:      char,
					modifiers: modifiers,
					pressed:   false,
				}
				return true
			},
			OnMouseButtonDown: func(buttonID uint8, x int32, y int32) bool {
				s.inputChan <- MouseInput{
					button:  buttonID,
					pressed: false,
					x:       x,
					y:       y,
				}
				return true
			},
			OnMouseButtonUp: func(buttonID uint8, x int32, y int32) bool {
				s.inputChan <- MouseInput{
					button:  buttonID,
					pressed: true,
					x:       x,
					y:       y,
				}
				return true
			},
		},
	})
	s.GameContainer.Focus()
	s.Client.RootWindow.AdoptChannel <- s.GameContainer.This

	// Sub-window: map
	err = s.MapElement.Setup(ui.MapElementConfig{
		Style: `
			X 50%
			Y 50%
			W 100%
			H 100%
			BackgroundColor 0 0 0 255
			Origin CenterX CenterY
		`,
	})
	mapText := ui.NewTextElement(ui.TextElementConfig{
		Value: "Map",
	})
	s.MapElement.AdoptChannel <- mapText
	s.GameContainer.AdoptChannel <- s.MapElement.This
	// Sub-window: chat
	err = s.ChatWindow.Setup(ui.ContainerConfig{
		Value: "Chat",
		Style: `
			X 8
			Y 8
			W 70%
			H 20%
			BackgroundColor 0 0 128 128
		`,
	})
	s.GameContainer.AdoptChannel <- s.ChatWindow.This
	// Sub-window: inventory
	err = s.InventoryWindow.Setup(ui.ContainerConfig{
		Value: "Inventory",
		Style: `
			X 50%
			Y 50%
			W 50%
			H 80%
			Origin CenterX CenterY
			BackgroundColor 0 128 0 128
		`,
	})
	s.GameContainer.AdoptChannel <- s.InventoryWindow.This
	s.InventoryWindow.SetHidden(true)
	// Sub-window: ground
	err = s.GroundWindow.Setup(ui.ContainerConfig{
		Value: "Ground",
		Style: `
			Y 70%
			W 30%
			H 30%
			BackgroundColor 128 128 128 128
		`,
	})
	s.GameContainer.AdoptChannel <- s.GroundWindow.This
	// Sub-window: stats
	err = s.StatsWindow.Setup(ui.ContainerConfig{
		Value: "Stats",
		Style: `
			X 30%
			W 40%
			H 20%
			BackgroundColor 128 0 0 128
		`,
	})
	s.GameContainer.AdoptChannel <- s.StatsWindow.This
	s.StatsWindow.SetHidden(true)
	// Sub-window: state
	err = s.StateWindow.Setup(ui.ContainerConfig{
		Value: "State",
		Style: `
			X 30%
			Y 80%
			W 40%
			H 20%
			BackgroundColor 128 128 0 128
		`,
	})
	s.GameContainer.AdoptChannel <- s.StateWindow.This
	s.StateWindow.SetHidden(true)
	//
	//go s.Client.LoopCmd()
	go s.Loop()
	return
}

// Close our Game state.
func (s *Game) Close() {
	s.MapElement.Destroy()
	s.StateWindow.Destroy()
	s.StatsWindow.Destroy()
	s.GroundWindow.Destroy()
	s.InventoryWindow.Destroy()
	s.ChatWindow.Destroy()
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
					}
				}
			case MouseInput:
				s.Client.Log.Printf("mouse: %+v\n", e)
			}
		}
		// TODO: For each object, create a corresponding ImageElement. These should then have their X,Y,Z set to their position based upon which Tile they exist in. Additionally, their Image would be synchronized to the object's current animation and face (as well as frame). It may be necessary to introduce Z-ordering, for both objects within the same tile, as well as for objects which exist at a higher Y.
		m := s.world.GetCurrentMap()
		if m == nil {
			continue
		}
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
	default:
		s.Client.Log.Printf("Server sent a Command %+v\n", c)
	}
	return false
}
