// +build !MOBILE
package UI

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"path"
)

type Instance struct {
	HeldElement     ElementI
	FocusedElement  ElementI
	HoveredElements []ElementI
	Running         bool
	RootWindow      Window
	Context         Context
}

func NewInstance() (inst *Instance, e error) {
	inst = &Instance{}
	return
}

func (i *Instance) Setup(data_root string) (err error) {
	// Initialize SDL
	if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	// Initialize TTF
	if err = ttf.Init(); err != nil {
		panic(err)
	}
	// Set up our UI Context
	if i.Context.Font, err = ttf.OpenFont(path.Join(data_root, "fonts", "DefaultFont.ttf"), 12); err != nil {
		panic(err)
	}

	err = i.RootWindow.Setup(WindowConfig{
		Value: "Chimera",
		Style: `
			BackgroundColor 0 0 0 255
			W 1280
			H 720
		`,
		RenderFunc: func(w *Window) {
			w.Context.Renderer.Clear()
		},
		Context: &i.Context,
	})
	return
}
func (i *Instance) Cleanup() {
	i.RootWindow.Destroy()
	sdl.Quit()
}

func (i *Instance) Loop() {
	i.Running = true
	// Render initial view.
	i.RootWindow.RenderMutex.Lock()
	i.RootWindow.Render()
	i.RootWindow.RenderMutex.Unlock()

	for i.Running {
		event := sdl.WaitEvent()
		switch t := event.(type) {
		case *sdl.QuitEvent:
			i.Running = false
		case *sdl.WindowEvent:
			if t.Event == sdl.WINDOWEVENT_RESIZED {
				i.RootWindow.RenderMutex.Lock()
				i.RootWindow.Resize(t.WindowID, t.Data1, t.Data2)
				i.RootWindow.RenderMutex.Unlock()
			} else if t.Event == sdl.WINDOWEVENT_CLOSE {
				i.Running = false
			} else if t.Event == sdl.WINDOWEVENT_EXPOSED {
				i.RootWindow.RenderMutex.Lock()
				i.RootWindow.Render()
				i.RootWindow.RenderMutex.Unlock()
			}
		default:
			i.HandleEvent(event)
			if i.RootWindow.HasDirt() {
				i.RootWindow.RenderMutex.Lock()
				i.RootWindow.Render()
				i.RootWindow.RenderMutex.Unlock()
			}
		}
	}
}

func (i *Instance) HandleEvent(event sdl.Event) {
	switch t := event.(type) {
	case *sdl.WindowEvent:
	case *sdl.MouseMotionEvent:
	case *sdl.MouseButtonEvent:
		if i.FocusedElement != nil {
			if !i.FocusedElement.Hit(t.X, t.Y) {
				if t.State == 1 {
					i.BlurFocusedElement()
				}
			}
		}
		if i.HeldElement != nil {
			if t.State == sdl.RELEASED && t.Button == sdl.BUTTON_LEFT {
				i.HeldElement.SetHeld(false)
				i.HeldElement = nil
			}
		}
	case *sdl.KeyboardEvent:
		if i.FocusedElement != nil {
			if t.Keysym.Sym == 27 {
				i.BlurFocusedElement()
				return
			} else if t.Keysym.Sym == 9 && t.State == sdl.RELEASED { // tab
				if t.Keysym.Mod&1 == 1 { // Shift
					i.FocusPreviousElement(i.FocusedElement)
				} else {
					i.FocusNextElement(i.FocusedElement)
				}
				return
			}
			if t.State == sdl.PRESSED {
				i.FocusedElement.OnKeyDown(uint8(t.Keysym.Sym), t.Keysym.Mod)
			} else {
				i.FocusedElement.OnKeyUp(uint8(t.Keysym.Sym), t.Keysym.Mod)
			}
			return
		}
	case *sdl.TextInputEvent:
		if i.FocusedElement != nil {
			i.FocusedElement.OnTextInput(t.GetText())
		}
		return
	case *sdl.TextEditingEvent:
		if i.FocusedElement != nil {
			i.FocusedElement.OnTextEdit(t.GetText(), t.Start, t.Length)
		}
		return
	}
	// If any events weren't handled above, we send the event down the tree.
	i.IterateEvent(i.RootWindow.This, event)
}

func (inst *Instance) IterateEvent(e ElementI, event sdl.Event) {
	switch t := event.(type) {
	case *sdl.WindowEvent:
	case *sdl.MouseMotionEvent:
		if e.Hit(t.X, t.Y) {
			// OnMouseIn
			existsInHovered := false
			for _, he := range inst.HoveredElements {
				if he == e {
					existsInHovered = true
					break
				}
			}
			if !existsInHovered {
				inst.HoveredElements = append(inst.HoveredElements, e)
				e.OnMouseIn(t.X, t.Y)
			}
			// OnMouseMove
			if !e.OnMouseMove(t.X, t.Y) {
				return
			}
		} else {
			// OnMouseOut
			for i, he := range inst.HoveredElements {
				if he == e {
					he.OnMouseOut(t.X, t.Y)
					inst.HoveredElements[i] = inst.HoveredElements[len(inst.HoveredElements)-1]
					inst.HoveredElements = inst.HoveredElements[:len(inst.HoveredElements)-1]
					break
				}
			}
		}
	case *sdl.MouseButtonEvent:
		if e.Hit(t.X, t.Y) {
			if t.State == sdl.PRESSED {
				if e.CanFocus() {
					inst.FocusElement(e)
				}
				if t.Button == sdl.BUTTON_LEFT && e.CanHold() {
					inst.HeldElement = e
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
		inst.IterateEvent(child, event)
	}
}

func (inst *Instance) BlurFocusedElement() {
	if inst.FocusedElement != nil {
		inst.FocusedElement.SetFocused(false)
		inst.FocusedElement.OnBlur()
	}
	inst.FocusedElement = nil
}

func (inst *Instance) FocusElement(e ElementI) {
	if inst.FocusedElement != nil && inst.FocusedElement != e {
		inst.FocusedElement.SetFocused(false)
		inst.FocusedElement.OnBlur()
	}
	e.SetFocused(true)
	e.OnFocus()
	inst.FocusedElement = e
}

func (inst *Instance) FocusNextElement(start ElementI) {
	found := false
	for _, c := range start.GetParent().GetChildren() {
		if c == start {
			found = true
		} else if found {
			if c.CanFocus() {
				inst.FocusElement(c)
				return
			}
		}
	}
	// if we get here just Blur the focused one
	inst.BlurFocusedElement()
}
func (inst *Instance) FocusPreviousElement(start ElementI) {
	found := false
	children := start.GetParent().GetChildren()
	for i := len(children) - 1; i >= 0; i-- {
		c := children[i]
		if c == start {
			found = true
		} else if found {
			if c.CanFocus() {
				inst.FocusElement(c)
				return
			}
		}
	}
	// if we get here just Blur the focused one
	inst.BlurFocusedElement()
}
