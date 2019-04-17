// +build !MOBILE
package UI

import (
	"github.com/veandco/go-sdl2/sdl"
	"sync"
)

type RenderFunc func(*Window)

type WindowConfig struct {
	Parent     *Window
	Style      Style
	RenderFunc RenderFunc
	Context    *Context
	Value      string
}

type Window struct {
	BaseElement
	SDL_window  *sdl.Window
	SDL_texture *sdl.Texture

	RenderFunc  RenderFunc
	RenderMutex sync.Mutex
}

func NewWindow(c WindowConfig) (w *Window, err error) {
	window := Window{}
	err = window.Setup(c)
	return &window, err
}

func (w *Window) Setup(c WindowConfig) (err error) {
	w.This = ElementI(w)
	w.RenderMutex = sync.Mutex{}
	w.RenderFunc = c.RenderFunc
	w.Style.Set(c.Style)
	w.Context = c.Context
	w.Value = c.Value
	if c.Parent != nil {
		w.SDL_window = c.Parent.SDL_window
		// NOTE: AdoptChild calls CalculateStyle
		c.Parent.AdoptChild(w)
	} else {
		w.SDL_window, err = sdl.CreateWindow(c.Value, int32(w.Style.X.Value), int32(w.Style.Y.Value), int32(w.Style.W.Value), int32(w.Style.H.Value), sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)
	}
	if err != nil {
		return err
	}
	// Create our Renderer
	w.Context.Renderer, err = w.SDL_window.GetRenderer()
	if w.Context.Renderer == nil {
		w.Context.Renderer, err = sdl.CreateRenderer(w.SDL_window, -1, sdl.RENDERER_ACCELERATED)
		w.Context.Renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
	}
	if err != nil {
		return err
	}
	w.CalculateStyle()
	// Trigger a resize so we can create a Texture
	//wid, err := w.SDL_window.GetID()
	//w.Resize(wid, w.w, w.h)
	if err != nil {
		return err
	}
	return nil
}

func (w *Window) Resize(id uint32, width int32, height int32) (err error) {
	wid, err := w.SDL_window.GetID()
	if wid == id {
		if w.Parent == nil {
			w.Style.W.Set(float64(width))
			w.Style.H.Set(float64(height))
			w.CalculateStyle()
		} else {
			w.CalculateStyle()
		}
	}
	return nil
}

func (w *Window) UpdateTexture() (err error) {
	if w.Parent == nil {
		return
	}
	var tw, th int32 = 0, 0
	if w.SDL_texture != nil {
		_, _, tw, th, err = w.SDL_texture.Query()
		if err != nil {
			panic(err)
		}
	}

	t, err := w.Context.Renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, w.w, w.h)
	t.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		return err
	}
	if w.SDL_texture != nil {
		w.Context.Renderer.SetRenderTarget(t)
		w.Context.Renderer.Copy(w.SDL_texture, nil, &sdl.Rect{0, 0, tw, th})
		w.SDL_texture.Destroy()
	}
	w.SDL_texture = t
	return
}

func (w *Window) Render() {
	if w.IsHidden() {
		return
	}
	old_t := w.Context.Renderer.GetRenderTarget()
	w.Context.Renderer.SetRenderTarget(w.SDL_texture)
	if w.RenderFunc != nil {
		w.RenderFunc(w)
	}
	w.BaseElement.Render()
	if w.Parent != nil {
		w.Context.Renderer.SetRenderTarget(old_t)
		w.Context.Renderer.Copy(w.SDL_texture, nil, &sdl.Rect{w.x, w.y, w.w, w.h})
	} else {
		w.Context.Renderer.Present()
	}
}

func (w *Window) CalculateStyle() {
	w.BaseElement.CalculateStyle()
	if w.IsDirty() {
		w.UpdateTexture()
	}
}
func (w *Window) Destroy() {
	if w.Parent == nil {
		w.SDL_window.Destroy()
		w.Context.Renderer.Destroy()
	} else {
		w.Parent.DisownChild(w)
	}
	if w.SDL_texture != nil {
		w.SDL_texture.Destroy()
	}
	for _, child := range w.Children {
		child.Destroy()
	}
}

func (w *Window) HandleEvent(event sdl.Event) {
	IterateEvent(w, event)
}

func IterateEvent(e ElementI, event sdl.Event) {
	switch t := event.(type) {
	case *sdl.WindowEvent:
	case *sdl.MouseMotionEvent:
		if e.Hit(t.X, t.Y) {
			if !e.OnMouseMove(t.X, t.Y) {
				return
			}
		}
	case *sdl.MouseButtonEvent:
		if e.Hit(t.X, t.Y) {
			if t.State == 1 {
				if !e.OnMouseButtonDown(t.Button, t.X, t.Y) {
					return
				}
			} else {
				if !e.OnMouseButtonUp(t.Button, t.X, t.Y) {
					return
				}
			}
		}
	case *sdl.KeyboardEvent:
		// ??? Probably have a global and currently focused element
	case *sdl.TextInputEvent:
		// This should only receive when a focused element is true and inputting
	case *sdl.TextEditingEvent:
		// This should somehow show a "temp" value in input elements.
	}
	for _, child := range e.GetChildren() {
		switch t := event.(type) {
		case *sdl.WindowEvent:
		case *sdl.MouseMotionEvent:
			if child.Hit(t.X, t.Y) {
				if !child.OnMouseMove(t.X, t.Y) {
					return
				}
			}
		case *sdl.MouseButtonEvent:
			if child.Hit(t.X, t.Y) {
				if t.State == 1 {
					if !child.OnMouseButtonDown(t.Button, t.X, t.Y) {
						return
					}
				} else {
					if !child.OnMouseButtonUp(t.Button, t.X, t.Y) {
						return
					}
				}
			}
		default:
		}
		IterateEvent(child, event)
	}
}

func (w *Window) GetX() int32 {
	return 0
}
func (w *Window) GetY() int32 {
	return 0
}
