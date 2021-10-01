package states

import (
	"fmt"

	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-common/network"
)

// UserInput is an interface used in a channel in Game for handling UI input.
type UserInput interface {
}

// ResizeEvent is used to notify the UI of a resize change.
type ResizeEvent struct{}

// ChatEvent is used to send an input chat to the main loop.
type ChatEvent struct {
	Body string
}

// DisconnectEvent is used to tell the client to disconnect.
type DisconnectEvent struct{}

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

// SetupUI sets up all the UI windows.
func (s *Game) SetupUI() (err error) {
	// Main Container
	err = s.GameContainer.Setup(ui.ContainerConfig{
		Value: "Game",
		Style: GameContainerStyle,
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
		Style: MapContainerStyle,
	})
	mapText := ui.NewTextElement(ui.TextElementConfig{
		Value: "Map",
	})
	s.MapContainer.AdoptChannel <- mapText
	s.GameContainer.AdoptChannel <- s.MapContainer.This

	// Sub-window: chat
	s.ChatWindow.Setup(ui.ContainerConfig{
		Style: ChatWindowStyle,
		Events: ui.Events{
			OnWindowResized: func(w, h int32) {
				s.inputChan <- ResizeEvent{}
			},
		},
	})

	err = s.MessagesWindow.Setup(ui.ContainerConfig{
		Value: "Messages",
		Style: MessagesWindowStyle,
		Events: ui.Events{
			OnWindowResized: func(w, h int32) {
				s.inputChan <- ResizeEvent{}
			},
		},
	})

	s.ChatInput = ui.NewInputElement(ui.InputElementConfig{
		Value:         "",
		Style:         ChatInputStyle,
		SubmitOnEnter: true,
		ClearOnSubmit: true,
		BlurOnSubmit:  true,
		Placeholder:   "...",
		Events: ui.Events{
			OnTextSubmit: func(str string) bool {
				if str == "" {
					return true
				}
				s.inputChan <- ChatEvent{
					Body: str,
				}
				return true
			},
			OnFocus: func() bool {
				s.ChatInput.GetStyle().BackgroundColor.A = 128
				return true
			},
			OnBlur: func() bool {
				s.ChatInput.GetStyle().BackgroundColor.A = 32
				return true
			},
		},
	})

	s.ChatWindow.GetAdoptChannel() <- s.MessagesWindow.This
	s.ChatWindow.GetAdoptChannel() <- s.ChatInput
	s.GameContainer.AdoptChannel <- s.ChatWindow.This
	// Sub-window: inventory
	err = s.InventoryWindow.Setup(ui.ContainerConfig{
		Value: "Inventory",
		Style: InventoryWindowStyle,
	})
	s.GameContainer.AdoptChannel <- s.InventoryWindow.This
	// Sub-window: ground
	err = s.GroundWindow.Setup(ui.ContainerConfig{
		Value: "Ground",
		Style: GroundWindowStyle,
	})
	s.GameContainer.AdoptChannel <- s.GroundWindow.This
	// Sub-window: stats
	err = s.StatsWindow.Setup(ui.ContainerConfig{
		Value: "Stats",
		Style: StatsWindowStyle,
	})
	s.GameContainer.AdoptChannel <- s.StatsWindow.This
	// Sub-window: state
	err = s.StateWindow.Setup(ui.ContainerConfig{
		Value: "State",
		Style: StateWindowStyle,
	})
	s.GameContainer.AdoptChannel <- s.StateWindow.This
	s.StateWindow.SetHidden(true)

	return err
}

// CleanupUI destroys all UI elements.
func (s *Game) CleanupUI() {
	s.MapContainer.Destroy()
	s.StateWindow.Destroy()
	s.StatsWindow.Destroy()
	s.GroundWindow.Destroy()
	s.InventoryWindow.Destroy()
	s.MessagesWindow.Destroy()
}

// UpdateMessagesWindow synchronizes the message window with the client's message history.
func (s *Game) UpdateMessagesWindow() {
	addMessage := func(str string) ui.ElementI {
		e := ui.NewTextElement(ui.TextElementConfig{
			Value: str,
			Style: fmt.Sprintf(`
				ForegroundColor 200 200 200 255
				OutlineColor 20 20 20 255
			`),
		})
		s.messageElements = append(s.messageElements, e)
		s.MessagesWindow.GetAdoptChannel() <- s.messageElements[len(s.messageElements)-1]
		return e
	}

	// Create message UI as needed.
	for i := len(s.Client.MessageHistory) - 1; i >= 0; i-- {
		if i >= len(s.messageElements) {
			m := s.Client.MessageHistory[i]
			msgName := ""
			if m.Message.Type == network.ServerMessage {
				msgName = "SERVER"
				addMessage(fmt.Sprintf("[%s] <%s>: %s", msgName, m.Received.Local(), m.Message.Body))
			} else if m.Message.Type == network.ChatMessage {
				msgName = "CHAT"
				addMessage(fmt.Sprintf("[%s] %s: %s", msgName, m.Message.From, m.Message.Body))
			} else if m.Message.Type == network.MapMessage {
				msgName = "MAP"
				addMessage(fmt.Sprintf("[%s] %s", msgName, m.Message.Body))
			} else if m.Message.Type == network.LocalMessage {
				addMessage(m.Message.Body)
			}
		}
	}
}
