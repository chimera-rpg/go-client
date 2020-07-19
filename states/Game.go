package states

import (
	"fmt"

	"github.com/chimera-rpg/go-client/client"
	"github.com/chimera-rpg/go-client/data"
	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-client/world"
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-common/network"
)

// Game is our live Game state, used once the user has connected to the server
// and joined as a player character.
type Game struct {
	client.State
	GameContainer   ui.Container
	ChatWindow      ui.Container
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
	s.objectImages = make(map[uint32]ui.ElementI)
	s.objectImageIDs = make(map[uint32]data.StringID)
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
	err = s.MapContainer.Setup(ui.ContainerConfig{
		Style: `
			X 50%
			Y 50%
			W 80%
			H 80%
			BackgroundColor 0 0 0 255
			Origin CenterX CenterY
		`,
	})
	mapText := ui.NewTextElement(ui.TextElementConfig{
		Value: "Map",
	})
	s.MapContainer.AdoptChannel <- mapText
	s.GameContainer.AdoptChannel <- s.MapContainer.This
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
	s.MapContainer.Destroy()
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
	default:
		s.Client.Log.Printf("Server sent a Command %+v\n", c)
	}
	return false
}

// HandleRender handles the rendering of our Game state.
func (s *Game) HandleRender() {
	// FIXME: This is _very_ rough and is just for testing!
	m := s.world.GetCurrentMap()
	objects := s.world.GetObjects()
	// Delete images that no longer correspond to an existing world object.
	for oID, t := range s.objectImages {
		o := s.world.GetObject(oID)
		if o == nil {
			t.GetDestroyChannel() <- true
			delete(s.objectImages, oID)
		}
	}
	// Iterate over world objects.
	for _, o := range objects {
		s.RenderObject(o, m)
	}
	return
}

func (s *Game) RenderObject(o *world.Object, m *world.DynamicMap) {
	scale := 4
	tileWidth := int(s.Client.AnimationsConfig.TileWidth)
	tileHeight := int(s.Client.AnimationsConfig.TileHeight)
	xOffset, yOffset := 0, 0
	// If the object is missing (out of view), delete it. FIXME: This should probably convert the image rendering to semi-opaque or otherwise instead.
	if o.Missing {
		if t, ok := s.objectImages[o.ID]; ok {
			t.GetDestroyChannel() <- true
			delete(s.objectImages, o.ID)
			delete(s.objectImageIDs, o.ID)
		}
		return
	}
	frames := s.Client.DataManager.GetFace(o.AnimationID, o.FaceID)
	// Bail if there are no frames to render.
	if len(frames) == 0 {
		return
	}
	// Calculate x and y render offset.
	if adjust, ok := s.Client.AnimationsConfig.Adjustments[cdata.ArchetypeType(o.Type)]; ok {
		xOffset += int(adjust.X)
		yOffset += int(adjust.Y)
	}

	xOffset += int(o.Y) * int(s.Client.AnimationsConfig.YStep.X)
	yOffset += int(o.Y) * int(-s.Client.AnimationsConfig.YStep.Y)

	startX := 0
	startY := int(m.GetHeight()) * int(-s.Client.AnimationsConfig.YStep.Y)

	oX := (int(o.X)*tileWidth + xOffset + startX)
	oY := (int(o.Z)*tileHeight - yOffset + startY)

	indexZ := int(o.Z)
	indexX := int(o.X)
	indexY := int(o.Y)

	zIndex := (indexZ * int(m.GetHeight()) * int(m.GetWidth())) + (int(m.GetDepth()) * indexY) - (indexX) + o.Index

	x := oX*scale + 100
	y := oY*scale + 100
	w := tileWidth * scale
	h := tileHeight * scale

	img := s.Client.DataManager.GetCachedImage(frames[0].ImageID)
	if _, ok := s.objectImages[o.ID]; !ok {
		if img != nil {
			bounds := img.Bounds()
			w = bounds.Max.X * scale
			h = bounds.Max.Y * scale
			if o.D > 1 {
				y -= h
				y += (int(o.D/2)*tileHeight + tileHeight/4) * scale
			}
			s.objectImages[o.ID] = ui.NewImageElement(ui.ImageElementConfig{
				Style: fmt.Sprintf(`
							X %d
							Y %d
							W %d
							H %d
							ZIndex %d
						`, x, y, w, h, zIndex),
				Image: img,
			})
			s.objectImageIDs[o.ID] = frames[0].ImageID
		} else {
			s.objectImages[o.ID] = ui.NewImageElement(ui.ImageElementConfig{
				Style: fmt.Sprintf(`
							X %d
							Y %d
							W %d
							H %d
							ZIndex %d
						`, x, y, w, h, zIndex),
				Image: img,
			})
		}
		s.MapContainer.GetAdoptChannel() <- s.objectImages[o.ID]
	} else {
		if img != nil {
			bounds := img.Bounds()
			w = bounds.Max.X * scale
			h = bounds.Max.Y * scale
			if o.D > 1 {
				y -= h
				y += (int(o.D/2)*tileHeight + tileHeight/4) * scale
			}
			if o.Changed {
				s.objectImages[o.ID].GetUpdateChannel() <- ui.UpdateX{ui.Number{Value: float64(x)}}
				s.objectImages[o.ID].GetUpdateChannel() <- ui.UpdateY{ui.Number{Value: float64(y)}}
				s.objectImages[o.ID].GetUpdateChannel() <- ui.UpdateW{ui.Number{Value: float64(w)}}
				s.objectImages[o.ID].GetUpdateChannel() <- ui.UpdateH{ui.Number{Value: float64(h)}}
				s.objectImages[o.ID].GetUpdateChannel() <- ui.UpdateZIndex{ui.Number{Value: float64(zIndex)}}
				o.Changed = false
			}
			// Only update the image if the image ID has changed.
			if s.objectImageIDs[o.ID] != frames[0].ImageID {
				s.objectImageIDs[o.ID] = frames[0].ImageID
				s.objectImages[o.ID].GetUpdateChannel() <- img
			}
		}
	}
}
