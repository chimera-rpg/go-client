// +build !MOBILE

package ui

import (
	"github.com/veandco/go-sdl2/sdl"
)

// Window is a UI element that can either represent a standalone OS window or a
// sub-window contained within another Window.
type Window struct {
	BaseElement
	SDLWindow *sdl.Window

	RenderFunc RenderFunc
}

// Setup our window object according to the passed WindowConfig.
func (w *Window) Setup(c WindowConfig) (err error) {
	w.This = ElementI(w)
	w.SetupChannels()
	w.RenderFunc = c.RenderFunc
	w.Style.Parse(WindowElementStyle)
	w.Style.Parse(c.Style)
	w.Context = c.Context
	w.Value = c.Value
	w.SetDirty(true)
	w.SDLWindow, err = sdl.CreateWindow(
		c.Value,
		int32(w.Style.X.Value),
		int32(w.Style.Y.Value),
		int32(w.Style.W.Value),
		int32(w.Style.H.Value),
		sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE,
	)

	if err != nil {
		return err
	}
	// Create our Renderer
	w.Context.Renderer, err = sdl.CreateRenderer(w.SDLWindow, -1, sdl.RENDERER_ACCELERATED)
	w.Context.Renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
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
		w.Style.W.Set(float64(width))
		w.Style.H.Set(float64(height))
		w.CalculateStyle()
	}
	return nil
}

// Render the window, its renderer function, and its children to its texture,
// thereafter rendering its texture to a Parent if it exists or to the screen
// if it is a top-level window.
func (w *Window) Render() {
	if w.IsHidden() {
		return
	}

	w.Context.Renderer.SetRenderTarget(nil)
	if w.RenderFunc != nil {
		w.RenderFunc(w)
	}
	if w.Style.BackgroundColor.A > 0 {
		w.Context.Renderer.SetDrawColor(
			w.Style.BackgroundColor.R,
			w.Style.BackgroundColor.G,
			w.Style.BackgroundColor.B,
			w.Style.BackgroundColor.A,
		)
		w.Context.Renderer.Clear()
	}

	w.BaseElement.Render()

	w.Context.Renderer.Present()
}

// Destroy the window, clearing the SDL context and destroying the SDLWindow if it is a top-level window.
func (w *Window) Destroy() {
	for _, child := range w.Children {
		child.Destroy()
	}

	w.SDLWindow.Destroy()
	w.Context.Renderer.Destroy()
}
