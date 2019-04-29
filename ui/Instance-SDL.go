// +build !MOBILE

package ui

import (
	"fmt"

	"github.com/chimera-rpg/go-client/data"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// Instance is the managing instance of the entire UI system.
type Instance struct {
	dataManager     *data.Manager
	HeldElement     ElementI
	FocusedElement  ElementI
	HoveredElements []ElementI
	Running         bool
	RootWindow      Window
	Context         Context
}

// GlobalInstance is our pointer to the GlobalInstance. Used for Focus/Blur
// calls from within Elements.
var GlobalInstance *Instance

// Setup sets up the needed libraries and pulls all needed data from the
// location passed in the call.
func (instance *Instance) Setup(dataManager *data.Manager) (err error) {
	instance.dataManager = dataManager
	// Initialize SDL
	if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return err
	}
	// Initialize TTF
	if err = ttf.Init(); err != nil {
		return err
	}
	// Set up our UI Context
	if instance.Context.Font, err = ttf.OpenFont(dataManager.GetDataPath("fonts", "DefaultFont.ttf"), 12); err != nil {
		return err
	}

	err = instance.RootWindow.Setup(WindowConfig{
		Value: "Chimera",
		Style: `
			BackgroundColor 0 0 0 255
			W 1280
			H 720
		`,
		RenderFunc: func(w *Window) {
			w.Context.Renderer.Clear()
		},
		Context: &instance.Context,
	})
	return
}

// Cleanup cleans up after our instance.
func (instance *Instance) Cleanup() {
	instance.RootWindow.Destroy()
	sdl.Quit()
}

// Loop is our main event handling and rendering loop.
func (instance *Instance) Loop() {
	instance.Running = true
	// Render initial view.
	instance.RootWindow.Render()
	for instance.Running {
		event := sdl.WaitEvent()
		switch t := event.(type) {
		case *sdl.QuitEvent:
			instance.Running = false
		case *sdl.WindowEvent:
			if t.Event == sdl.WINDOWEVENT_RESIZED {
				instance.RootWindow.Resize(t.WindowID, t.Data1, t.Data2)
			} else if t.Event == sdl.WINDOWEVENT_CLOSE {
				instance.Running = false
			} else if t.Event == sdl.WINDOWEVENT_EXPOSED {
				instance.RootWindow.Render()
			}
		default:
			instance.HandleEvent(event)
			if instance.RootWindow.HasDirt() {
				instance.RootWindow.Render()
			}
		}
	}
}

// HandleEvent handles the passed SDL events from Loop.
func (instance *Instance) HandleEvent(event sdl.Event) {
	switch t := event.(type) {
	case *sdl.WindowEvent:
	case *sdl.MouseMotionEvent:
	case *sdl.MouseButtonEvent:
		if instance.FocusedElement != nil {
			if !instance.FocusedElement.Hit(t.X, t.Y) {
				if t.State == 1 {
					instance.BlurFocusedElement()
				}
			}
		}
		if instance.HeldElement != nil {
			if t.State == sdl.RELEASED && t.Button == sdl.BUTTON_LEFT {
				instance.HeldElement.SetHeld(false)
				instance.HeldElement = nil
			}
		}
	case *sdl.KeyboardEvent:
		if instance.FocusedElement != nil {
			if t.Keysym.Sym == 27 {
				instance.BlurFocusedElement()
				return
			} else if t.Keysym.Sym == 9 && t.State == sdl.RELEASED { // tab
				if t.Keysym.Mod&1 == 1 { // Shift
					instance.FocusPreviousElement(instance.FocusedElement)
				} else {
					instance.FocusNextElement(instance.FocusedElement)
				}
				return
			}
			if t.State == sdl.PRESSED {
				instance.FocusedElement.OnKeyDown(uint8(t.Keysym.Sym), t.Keysym.Mod)
			} else {
				instance.FocusedElement.OnKeyUp(uint8(t.Keysym.Sym), t.Keysym.Mod)
			}
			return
		}
	case *sdl.TextInputEvent:
		if instance.FocusedElement != nil {
			instance.FocusedElement.OnTextInput(t.GetText())
		}
		return
	case *sdl.TextEditingEvent:
		if instance.FocusedElement != nil {
			instance.FocusedElement.OnTextEdit(t.GetText(), t.Start, t.Length)
		}
		return
	}
	// If any events weren't handled above, we send the event down the tree.
	instance.IterateEvent(instance.RootWindow.This, event)
}

// IterateEvent handles iterating an event down the entire Element tree
// starting at the passed element.
func (instance *Instance) IterateEvent(e ElementI, event sdl.Event) {
	switch t := event.(type) {
	case *sdl.WindowEvent:
	case *sdl.MouseMotionEvent:
		if e.Hit(t.X, t.Y) {
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
				e.OnMouseIn(t.X, t.Y)
			}
			// OnMouseMove
			if !e.OnMouseMove(t.X, t.Y) {
				return
			}
		} else {
			// OnMouseOut
			for i, he := range instance.HoveredElements {
				if he == e {
					he.OnMouseOut(t.X, t.Y)
					instance.HoveredElements[i] = instance.HoveredElements[len(instance.HoveredElements)-1]
					instance.HoveredElements = instance.HoveredElements[:len(instance.HoveredElements)-1]
					break
				}
			}
		}
	case *sdl.MouseButtonEvent:
		if e.Hit(t.X, t.Y) {
			if t.State == sdl.PRESSED {
				if e.CanFocus() {
					instance.FocusElement(e)
				}
				if t.Button == sdl.BUTTON_LEFT && e.CanHold() {
					instance.HeldElement = e
					e.SetHeld(true)
				}
				if !e.OnMouseButtonDown(t.Button, t.X, t.Y) {
					return
				}
			} else {
				if !e.OnMouseButtonUp(t.Button, t.X, t.Y) {
					return
				}
			}
		} else {
			if t.State == 1 {
				//BlurFocusedElement()
			}
		}
	case *sdl.KeyboardEvent:
		if t.State == sdl.PRESSED {
			if !e.OnKeyDown(uint8(t.Keysym.Sym), t.Keysym.Mod) {
				return
			}
		} else {
			if !e.OnKeyUp(uint8(t.Keysym.Sym), t.Keysym.Mod) {
				return
			}
		}
	}
	for _, child := range e.GetChildren() {
		instance.IterateEvent(child, event)
	}
}

// BlurFocusedElement blurs the current focused element if it exists.
func (instance *Instance) BlurFocusedElement() {
	if instance.FocusedElement != nil {
		instance.FocusedElement.SetFocused(false)
		instance.FocusedElement.OnBlur()
	}
	instance.FocusedElement = nil
}

// FocusElement focuses the target element, blurring the previous element if
// it exists.
func (instance *Instance) FocusElement(e ElementI) {
	if instance.FocusedElement != nil && instance.FocusedElement != e {
		instance.FocusedElement.SetFocused(false)
		instance.FocusedElement.OnBlur()
	}
	e.SetFocused(true)
	e.OnFocus()
	instance.FocusedElement = e
}

// FocusNextElement finds and focuses the next focusable element after
// the passed element.
func (instance *Instance) FocusNextElement(start ElementI) {
	found := false
	for _, c := range start.GetParent().GetChildren() {
		if c == start {
			found = true
		} else if found {
			if c.CanFocus() {
				instance.FocusElement(c)
				return
			}
		}
	}
	// if we get here just Blur the focused one
	instance.BlurFocusedElement()
}

// FocusPreviousElement finds and focuses the previous element before
// the passed element.
func (instance *Instance) FocusPreviousElement(start ElementI) {
	found := false
	children := start.GetParent().GetChildren()
	for i := len(children) - 1; i >= 0; i-- {
		c := children[i]
		if c == start {
			found = true
		} else if found {
			if c.CanFocus() {
				instance.FocusElement(c)
				return
			}
		}
	}
	// if we get here just Blur the focused one
	instance.BlurFocusedElement()
}

func showWindow(flags uint32, format string, a ...interface{}) {
	var win *sdl.Window

	buttons := []sdl.MessageBoxButtonData{
		{Flags: sdl.MESSAGEBOX_BUTTON_RETURNKEY_DEFAULT, ButtonID: 1, Text: "OH NO"},
	}

	if GlobalInstance != nil && GlobalInstance.RootWindow.SDLWindow != nil {
		win = GlobalInstance.RootWindow.SDLWindow
	}

	messageboxdata := sdl.MessageBoxData{
		Flags:       flags,
		Window:      win,
		Title:       "Chimera",
		Message:     fmt.Sprintf(format, a...),
		Buttons:     buttons,
		ColorScheme: nil,
	}

	sdl.ShowMessageBox(&messageboxdata)
}

// ShowError shows a popup error window.
func ShowError(format string, a ...interface{}) {
	showWindow(sdl.MESSAGEBOX_ERROR, format, a...)
}

// ShowWarning shows a popup warning window.
func ShowWarning(format string, a ...interface{}) {
	showWindow(sdl.MESSAGEBOX_WARNING, format, a...)
}

// ShowInfo shows a popup info window.
func ShowInfo(format string, a ...interface{}) {
	showWindow(sdl.MESSAGEBOX_INFORMATION, format, a...)
}
