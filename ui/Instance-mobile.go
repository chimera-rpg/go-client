// +build mobile

package ui

import (
	"fmt"

	"github.com/chimera-rpg/go-client/data"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
)

// Setup sets up the needed libraries and pulls all needed data from the
// location passed in the call.
func (instance *Instance) Setup(dataManager DataManagerI) (err error) {
	instance.dataManager = dataManager

	err = instance.RootWindow.Setup(WindowConfig{
		Value: "Chimera",
		Style: `
			BackgroundColor 0 0 0 255
			W 1280
			H 720
		`,
		Context: &instance.Context,
	})
	return
}

// Cleanup cleans up after our instance.
func (instance *Instance) Cleanup() {
	instance.RootWindow.Destroy()
}

// Loop is our main event handling and rendering loop. It runs at 60 frames
// per second.
func (instance *Instance) Loop() {
	instance.Running = true

	app.Main(func(a app.App) {
		var glctx gl.Context
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					// Get gl Context
					var ok bool
					if glctx, ok = e.DrawContext.(gl.Context); !ok {
						panic("Couldn't open GL Context")
					}
					// Set it to Context?
					instance.Context.GLContext = glctx
					fmt.Printf("Set context... %v+\n", instance.Context)
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					// We're just quitting here. We should only do this on StageAlive&CrossOff.
					instance.Running = false
					return
				}
			case size.Event:
				instance.Context.Width = e.WidthPx
				instance.Context.Height = e.HeightPx
				instance.Context.PixelRatio = float32(e.HeightPx) / float32(e.WidthPx)
				// Do something with size event
				// TODO: Invalidate root?
			case paint.Event:
				if glctx == nil || e.External {
					continue
				}
				// Seems as good of a palce as any to manage our elements
				instance.CheckChannels(instance.RootWindow.This)
				instance.Render()
				a.Publish()
				a.Send(paint.Event{})
			default: // touch.Event, key.Event
				instance.HandleEvent(e)
			}
		}
	})
	instance.Cleanup()
}

func (instance *Instance) Render() {
	if instance.Context.GLContext == nil {
		return
	}

	if instance.RootWindow.HasDirt() {
		instance.RootWindow.Render()
	}

	// ???
	instance.Context.GLContext.Enable(gl.DEPTH_TEST)
}

// HandleEvent handles the passed SDL events from Loop.
func (instance *Instance) HandleEvent(event interface{}) {

	switch t := event.(type) {
	case touch.Event:
		if instance.FocusedElement != nil {
			if !instance.FocusedElement.Hit(int32(t.X), int32(t.Y)) {
				instance.BlurFocusedElement()
			}
		}
		if instance.HeldElement != nil {
			if t.Type == touch.TypeEnd {
				instance.HeldElement.SetHeld(false)
				instance.HeldElement = nil
			}
		}
	case key.Event:
		if instance.FocusedElement != nil {
			if t.Code == key.CodeEscape {
				instance.BlurFocusedElement()
				return
			} else if t.Code == key.CodeTab && t.Direction == key.DirRelease {
				if (t.Modifiers & key.ModShift) == key.ModShift {
					instance.FocusPreviousElement(instance.FocusedElement)
				} else {
					instance.FocusNextElement(instance.FocusedElement)
				}
				return
			}
			if t.Direction == key.DirRelease {
				// TODO: Handle repeat event if that is a thing.
				instance.FocusedElement.OnKeyDown(uint8(t.Code), uint16(t.Modifiers), false)
			} else {
				instance.FocusedElement.OnKeyUp(uint8(t.Code), uint16(t.Modifiers))
			}
			return
		}
	}
	// If any events weren't handled above, we send the event down the tree.
	instance.IterateEvent(instance.RootWindow.This, event)
}

// IterateEvent handles iterating an event down the entire Element tree
// starting at the passed element.
func (instance *Instance) IterateEvent(e ElementI, event interface{}) {
	switch t := event.(type) {
	case touch.Event:
		if t.Type == touch.TypeMove {
			if e.Hit(int32(t.X), int32(t.Y)) {
				// OnMouseIn
				existsInHovered := false
				for _, he := range instance.HoveredElements {
					if he == e {
						existsInHovered = true
						break
					}
				}
				if !existsInHovered {
					instance.HoveredElements = append(instance.HoveredElements, e)
					e.OnMouseIn(int32(t.X), int32(t.Y))
				}
				// OnMouseMove
				if !e.OnMouseMove(int32(t.X), int32(t.Y)) {
					return
				}
			} else {
				// OnMouseOut
				for i, he := range instance.HoveredElements {
					if he == e {
						he.OnMouseOut(int32(t.X), int32(t.Y))
						instance.HoveredElements[i] = instance.HoveredElements[len(instance.HoveredElements)-1]
						instance.HoveredElements = instance.HoveredElements[:len(instance.HoveredElements)-1]
						break
					}
				}
			}
		} else {
			if e.Hit(int32(t.X), int32(t.Y)) {
				if t.Type == touch.TypeBegin {
					if e.CanFocus() {
						instance.FocusElement(e)
					}
					if e.CanHold() {
						instance.HeldElement = e
						e.SetHeld(true)
					}
					if !e.OnMouseButtonDown(0, int32(t.X), int32(t.Y)) {
						return
					}
				} else {
					if !e.OnMouseButtonUp(0, int32(t.X), int32(t.Y)) {
						return
					}
				}
			}
		}
	case key.Event:
		if t.Direction == key.DirPress {
			if !e.OnKeyDown(uint8(t.Code), uint16(t.Modifiers)) {
				return
			}
		} else if t.Direction == key.DirRelease {
			if !e.OnKeyUp(uint8(t.Code), uint16(t.Modifiers)) {
				return
			}
		} else if t.Direction == key.DirNone {

		}
	}
	for _, child := range e.GetChildren() {
		instance.IterateEvent(child, event)
	}
}

func showWindow(flags uint32, format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

// ShowError shows a popup error window.
func ShowError(format string, a ...interface{}) {
	showWindow(3, format, a...)
}

// ShowWarning shows a popup warning window.
func ShowWarning(format string, a ...interface{}) {
	showWindow(2, format, a...)
}

// ShowInfo shows a popup info window.
func ShowInfo(format string, a ...interface{}) {
	showWindow(1, format, a...)
}
