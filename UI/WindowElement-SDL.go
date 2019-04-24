// +build !MOBILE

package ui

import (
	"sync"

	"github.com/veandco/go-sdl2/sdl"
)

// RenderFunc is a type alias used for the custom render function callback.
type RenderFunc func(*Window)

// WindowConfig is a configuration object that is used by NewWindow(c) or
// Window.Setup(c) to construct a functioning Window.
type WindowConfig struct {
	Parent     *Window
	Style      string
	RenderFunc RenderFunc
	Context    *Context
	Value      string
}

// Window is a UI element that can either represent a standalone OS window or a
// sub-window contained within another Window.
type Window struct {
	BaseElement
	SDLWindow  *sdl.Window
	SDLTexture *sdl.Texture

	RenderFunc  RenderFunc
	RenderMutex sync.Mutex
}

// WindowElementStyle provides the default Style that is applied to all windows.
var WindowElementStyle = `
	ForegroundColor 0 0 0 255
	BackgroundColor 139 186 139 255
`

// NewWindow creates a new Window instance according to the passed WindowConfig.
func NewWindow(c WindowConfig) (w *Window, err error) {
	window := Window{}
	err = window.Setup(c)
	return &window, err
}

// Setup our window object according to the passed WindowConfig.
func (w *Window) Setup(c WindowConfig) (err error) {
	w.This = ElementI(w)
	w.RenderMutex = sync.Mutex{}
	w.RenderFunc = c.RenderFunc
	w.Style.Parse(WindowElementStyle)
	w.Style.Parse(c.Style)
	w.Context = c.Context
	w.Value = c.Value
	if c.Parent != nil {
		w.SDLWindow = c.Parent.SDLWindow
		// NOTE: AdoptChild calls CalculateStyle
		c.Parent.RenderMutex.Lock()
		c.Parent.AdoptChild(w)
		c.Parent.RenderMutex.Unlock()
	} else {
		w.SDLWindow, err = sdl.CreateWindow(c.Value, int32(w.Style.X.Value), int32(w.Style.Y.Value), int32(w.Style.W.Value), int32(w.Style.H.Value), sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)
	}
	if err != nil {
		return err
	}
	// Create our Renderer
	w.Context.Renderer, err = w.SDLWindow.GetRenderer()
	if w.Context.Renderer == nil {
		w.Context.Renderer, err = sdl.CreateRenderer(w.SDLWindow, -1, sdl.RENDERER_ACCELERATED)
		w.Context.Renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
	}
	if err != nil {
		return err
	}
	w.CalculateStyle()
	// Trigger a resize so we can create a Texture
	//wid, err := w.SDLWindow.GetID()
	//w.Resize(wid, w.w, w.h)
	if err != nil {
		return err
	}
	return nil
}

// Resize the given SDL Window to a specific width and height. Intended for top-level windows only.
func (w *Window) Resize(id uint32, width int32, height int32) (err error) {
	wid, err := w.SDLWindow.GetID()
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

func (w *Window) updateTexture() (err error) {
	if w.Parent == nil {
		return
	}
	var tw, th int32 = 0, 0
	if w.SDLTexture != nil {
		_, _, tw, th, err = w.SDLTexture.Query()
		if err != nil {
			panic(err)
		}
	}

	t, err := w.Context.Renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, w.w, w.h)
	t.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		return err
	}
	if w.SDLTexture != nil {
		w.Context.Renderer.SetRenderTarget(t)
		w.Context.Renderer.Copy(w.SDLTexture, nil, &sdl.Rect{0, 0, tw, th})
		w.SDLTexture.Destroy()
	}
	w.SDLTexture = t
	return
}

// Render the window, its renderer function, and its children to its texture,
// thereafter rendering its texture to a Parent if it exists or to the screen
// if it is a top-level window.
func (w *Window) Render() {
	if w.IsHidden() {
		return
	}
	oldTexture := w.Context.Renderer.GetRenderTarget()
	w.Context.Renderer.SetRenderTarget(w.SDLTexture)
	if w.RenderFunc != nil {
		w.RenderFunc(w)
	}
	if w.Style.BackgroundColor.A > 0 {
		w.Context.Renderer.SetDrawColor(w.Style.BackgroundColor.R, w.Style.BackgroundColor.G, w.Style.BackgroundColor.B, w.Style.BackgroundColor.A)
		w.Context.Renderer.Clear()
	}

	w.BaseElement.Render()
	if w.Parent != nil {
		w.Context.Renderer.SetRenderTarget(oldTexture)
		w.Context.Renderer.Copy(w.SDLTexture, nil, &sdl.Rect{w.x, w.y, w.w, w.h})
	} else {
		w.Context.Renderer.Present()
	}
}

// CalculateStyle recalculates the style and updates the Window texture if it is dirty. See BaseElement.CalculateStyle().
func (w *Window) CalculateStyle() {
	w.BaseElement.CalculateStyle()
	if w.IsDirty() {
		w.updateTexture()
	}
}

// Destroy the window, clearing the SDL context and destroying the SDLWindow if it is a top-level window.
func (w *Window) Destroy() {
	if w.Parent == nil {
		w.SDLWindow.Destroy()
		w.Context.Renderer.Destroy()
	} else {
		w.Parent.DisownChild(w)
	}
	if w.SDLTexture != nil {
		w.SDLTexture.Destroy()
	}
	for _, child := range w.Children {
		child.Destroy()
	}
}

// GetX returns the cached x property. In the case of Windows this is 0.
func (w *Window) GetX() int32 {
	return 0
}

// GetY returns the cached y property. In the case of Windows this is 0.
func (w *Window) GetY() int32 {
	return 0
}

// IsContainer Returns whether or not this Element should be considered as a container.
func (w *Window) IsContainer() bool {
	return true
}
