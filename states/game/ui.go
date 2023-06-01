package game

import (
	"fmt"
	"image/color"

	"github.com/chimera-rpg/go-client/states/game/elements"
	"github.com/chimera-rpg/go-client/ui"
	cdata "github.com/chimera-rpg/go-server/data"
	"github.com/chimera-rpg/go-server/network"
)

// ChangeCommandMode notifies the UI to change the command mode.
type ChangeCommandMode struct{}

// DisconnectEvent is used to tell the client to disconnect.
type DisconnectEvent struct{}

// KeyInput is the Userinput for key events.
type KeyInput struct {
	code      uint8
	modifiers uint16
	pressed   bool
	repeat    bool
}

type GroundCell struct {
	element ui.Container
	image   ui.ElementI
	text    ui.ElementI
}

// SetupUI sets up all the UI windows.
func (s *Game) SetupUI() (err error) {
	if s.Client.DataManager.Config.Game.Graphics.ObjectScale == 0 {
		s.Client.DataManager.Config.Game.Graphics.ObjectScale = 4
	}
	s.objectsScale = &s.Client.DataManager.Config.Game.Graphics.ObjectScale
	fmt.Println("objectsScale", *s.objectsScale)
	// Main Container
	err = s.GameContainer.Setup(ui.ContainerConfig{
		Value: "Game",
		Style: s.Styles()["Game"]["Container"],
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
		},
	})
	s.GameContainer.Focus()
	s.Client.RootWindow.AdoptChannel <- s.GameContainer.This

	container, err := s.MapWindow.Setup(s.Styles()["Game"]["Map"], s.inputChan)
	s.GameContainer.AdoptChannel <- container.This

	// Sub-window: chat
	s.ChatWindow.Setup(ui.ContainerConfig{
		Style: s.Styles()["Game"]["Chat"],
		Events: ui.Events{
			OnWindowResized: func(w, h int32) {
				s.inputChan <- elements.ResizeEvent{}
			},
		},
	})

	err = s.MessagesWindow.Setup(ui.ContainerConfig{
		Value: "Messages",
		Style: s.Styles()["Game"]["Messages"],
		Events: ui.Events{
			OnWindowResized: func(w, h int32) {
				s.inputChan <- elements.ResizeEvent{}
			},
		},
	})

	s.CommandContainer, err = ui.NewContainerElement(ui.ContainerConfig{
		Style: s.Styles()["Game"]["CommandContainer"],
	})

	s.ChatType = ui.NewButtonElement(ui.ButtonElementConfig{
		Value:   CommandModeStrings[s.CommandMode],
		Style:   s.Styles()["Game"]["CommandType"],
		NoFocus: true,
		NoHold:  true,
		Events: ui.Events{
			OnMouseButtonUp: func(button uint8, x, y int32) bool {
				s.inputChan <- ChangeCommandMode{}
				return false
			},
		},
	})

	s.ChatInput = ui.NewInputElement(ui.InputElementConfig{
		Value:         "",
		Style:         s.Styles()["Game"]["ChatInput"],
		SubmitOnEnter: true,
		ClearOnSubmit: true,
		BlurOnSubmit:  true,
		Placeholder:   "...",
		Events: ui.Events{
			OnTextSubmit: func(str string) bool {
				if str == "" {
					return true
				}
				s.inputChan <- elements.ChatEvent{
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
	s.ChatWindow.GetAdoptChannel() <- s.CommandContainer
	s.CommandContainer.GetAdoptChannel() <- s.ChatType
	s.CommandContainer.GetAdoptChannel() <- s.ChatInput
	s.GameContainer.AdoptChannel <- s.ChatWindow.This
	// Sub-window: inventory
	inventoryContainer, err := s.InventoryWindow.Setup(s, elements.ContainerWindowConfig{Style: s.Styles()["Game"]["Inventory"], Type: elements.ContainerInventoryType, ID: "inventory"}, s.inputChan)
	if err != nil {
		panic(err)
	}
	s.GameContainer.AdoptChannel <- inventoryContainer.This
	// Sub-window: inspector
	inspectorContainer, err := s.InspectorWindow.Setup(s, s.Styles()["Game"]["Inspector"], s.inputChan)
	if err != nil {
		panic(err)
	}
	s.GameContainer.AdoptChannel <- inspectorContainer.This

	// Sub-window: ground
	groundContainer, err := s.GroundWindow.Setup(s, elements.ContainerWindowConfig{Style: s.Styles()["Game"]["Ground"], Type: elements.ContainerGroundType, ID: "ground"}, s.inputChan)
	if err != nil {
		panic(err)
	}
	s.GameContainer.AdoptChannel <- groundContainer.This
	// Sub-window: debug
	debugContainer, err := s.DebugWindow.Setup(s, s.Styles()["Game"]["Debug"], s.inputChan)
	if err != nil {
		panic(err)
	}
	s.GameContainer.AdoptChannel <- debugContainer.This

	// Sub-window: stats
	err = s.StatsWindow.Setup(ui.ContainerConfig{
		Value: "Stats",
		Style: s.Styles()["Game"]["Stats"],
	})
	s.GameContainer.AdoptChannel <- s.StatsWindow.This
	// Sub-window: state
	err = s.StateWindow.Setup(ui.ContainerConfig{
		Value: "State",
		Style: s.Styles()["Game"]["State"],
	})
	s.GameContainer.AdoptChannel <- s.StateWindow.This

	s.focusedImage = ui.NewImageElement(ui.ImageElementConfig{
		HideImage: true,
		Style:     s.Styles()["Game"]["FocusedImage"],
		Events: ui.Events{
			OnChange: func() bool {
				if o := s.world.GetObject(s.focusedObjectID); o != nil && o.Element != nil {
					s.focusedImage.GetStyle().X = o.Element.GetStyle().X
					s.focusedImage.GetStyle().Y = o.Element.GetStyle().Y
					s.focusedImage.GetStyle().W = o.Element.GetStyle().W
					s.focusedImage.GetStyle().H = o.Element.GetStyle().H
					s.focusedImage.SetDirty(true)
				}
				return true
			},
		},
	})
	s.MapWindow.Container.AdoptChannel <- s.focusedImage

	return err
}

// CleanupUI destroys all UI elements.
func (s *Game) CleanupUI() {
	s.GameContainer.GetDestroyChannel() <- true
}

// UpdateMessagesWindow synchronizes the message window with the client's message history.
func (s *Game) UpdateMessagesWindow() {
	addMessage := func(str string) ui.ElementI {
		e := ui.NewTextElement(ui.TextElementConfig{
			Value: str,
			Style: s.Styles()["Game"]["GenericMessage"],
		})
		s.messageElements = append(s.messageElements, e)
		s.MessagesWindow.GetAdoptChannel() <- s.messageElements[len(s.messageElements)-1]
		return e
	}

	// Create message UI as needed.
	for i := len(s.MessageHistory) - 1; i >= 0; i-- {
		if i >= len(s.messageElements) {
			m := s.MessageHistory[i]
			msgName := ""
			// Just print server messages.
			if m.Message.Type == network.ServerMessage {
				msgName = "SERVER"
				addMessage(fmt.Sprintf("[%s] <%s>: %s", msgName, m.Received.Local(), m.Message.Body))
			} else if m.Message.Type == network.ChatMessage {
				// Just print chat messages.
				msgName = "CHAT"
				addMessage(fmt.Sprintf("[%s] %s: %s", msgName, m.Message.From, m.Message.Body))
			} else if m.Message.Type == network.TargetMessage {
				// Target messages get printed plainly.
				if m.Message.FromObjectID != s.world.GetViewObject().ID {
					n := "???"
					o := s.world.GetObject(m.Message.FromObjectID)
					if o != nil {
						// TODO: Look up object or something...?
					}
					addMessage(fmt.Sprintf("%s: %s", n, m.Message.Body))
				} else {
					addMessage(fmt.Sprintf("%s", m.Message.Body))
				}
			} else if m.Message.Type == network.NPCMessage || m.Message.Type == network.PCMessage {
				// NPC/PC messages print as `X says: ...` and provide either a truncated version of the statement as floating text or the msg Title as the floating text. If the object is not known no floating text is shown.
				// TODO: It'd be nice if we had a local objectID -> name field we could use.
				o := s.world.GetObject(m.Message.FromObjectID)
				if o != nil {
					col := color.RGBA{255, 255, 255, 255}
					if m.Message.Type == network.NPCMessage {
						col = color.RGBA{128, 128, 128, 200}
					} else if o == s.world.GetViewObject() {
						col = color.RGBA{255, 255, 255, 200}
					}
					// Prefer using the message's Title for the popup text.
					text := m.Message.Title
					if text == "" {
						if len(m.Message.Body) > 40 {
							text = m.Message.Body[:40] + "..."
						} else {
							text = m.Message.Body
						}
					}
					mapMessage, err := s.createMapObjectMessage(m.Message.FromObjectID, text, col)
					if err != nil {
						// TODO: Print some sort of error.
					}
					s.MapWindow.Messages = append(s.MapWindow.Messages, mapMessage)
					s.MapWindow.Container.GetAdoptChannel() <- mapMessage.El
				}
				// FIXME: Replace wtih GetPlayerObject()
				if o == s.world.GetViewObject() {
					addMessage(fmt.Sprintf("You speak: %s", m.Message.Body))
				} else {
					addMessage(fmt.Sprintf("%s speaks: %s", m.Message.From, m.Message.Body))
				}
			} else if m.Message.Type == network.MapMessage {
				msgName = "MAP"
				addMessage(fmt.Sprintf("[%s] %s", msgName, m.Message.Body))
			} else if m.Message.Type == network.LocalMessage {
				addMessage(m.Message.Body)
			}
		}
	}
}

func (s *Game) UpdateStateWindow() {
	addStatus := func(status cdata.StatusType) ui.ElementI {
		e := ui.NewTextElement(ui.TextElementConfig{
			Value: cdata.StatusMapToString[status],
			Style: s.Styles()["Game"]["State__Status"],
		})
		s.statusElements[status] = e
		s.StateWindow.GetAdoptChannel() <- e
		return e
	}

	// Add any missing.
	for k := range s.statuses {
		if _, ok := s.statusElements[k]; !ok {
			addStatus(k)
		}
	}

	// Readjust UI.
	y := int32(0)
	for k, v := range s.statusElements {
		v.GetUpdateChannel() <- ui.UpdateY{Number: ui.Number{Value: float64(y)}}
		if s.statuses[k] {
			v.GetUpdateChannel() <- ui.UpdateHidden(false)
		} else {
			v.GetUpdateChannel() <- ui.UpdateHidden(true)
		}
		y += v.GetHeight()
	}
}

func (s *Game) UpdateGroundWindow() {
	// This should show nearby items, organized by view object Y to view object Height, from bottom to top. This also has to take in mind width and depth, as well as reach, so this becomes difficult... perhaps there could be a toggle for view modes, one with each height, width, depth, stack, in columns, the other with "nearby items" that just show all non-block, non-tile, non-character, and similar items in a column.
}
