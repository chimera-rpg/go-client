//go:build !mobile
// +build !mobile

package ui

import (
	"github.com/veandco/go-sdl2/sdl"
)

// Container is a UI element that represents a texture-backed containing element.
type Container struct {
	BaseElement
	SDLWindow  *sdl.Window
	SDLTexture *sdl.Texture

	ContainerRenderFunc ContainerRenderFunc
}

func (w *Container) updateTexture() (err error) {
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
		w.Context.Renderer.Copy(w.SDLTexture, nil, &sdl.Rect{X: 0, Y: 0, W: tw, H: th})
		w.SDLTexture.Destroy()
	}
	w.SDLTexture = t
	return
}

// Render the window, its renderer function, and its children to its texture,
// thereafter rendering its texture to a Parent if it exists or to the screen
// if it is a top-level window.
func (w *Container) Render() {
	if w.IsHidden() {
		return
	}
	oldTexture := w.Context.Renderer.GetRenderTarget()
	w.Context.Renderer.SetRenderTarget(w.SDLTexture)
	if w.ContainerRenderFunc != nil {
		w.ContainerRenderFunc(w)
	}
	if w.Style.BackgroundColor.A > 0 {
		w.Context.Renderer.SetDrawColor(w.Style.BackgroundColor.R, w.Style.BackgroundColor.G, w.Style.BackgroundColor.B, w.Style.BackgroundColor.A)
	} else {
		w.Context.Renderer.SetDrawColor(0, 0, 0, 0)
	}
	w.Context.Renderer.Clear()

	w.BaseElement.Render()
	for _, child := range w.BaseElement.VisibleChildren() {
		child.RenderPost()
	}
	if w.Parent != nil {
		w.Context.Renderer.SetRenderTarget(oldTexture)
		w.Context.Renderer.Copy(w.SDLTexture, nil, &sdl.Rect{X: w.x, Y: w.y, W: w.w, H: w.h})
	} else {
		w.Context.Renderer.Present()
	}
}

// CalculateStyle recalculates the style and updates the Container texture if it is dirty. See BaseElement.CalculateStyle().
func (w *Container) CalculateStyle() {
	w.BaseElement.CalculateStyle()
	if w.IsDirty() {
		w.reflow()
		// Update texture.
		w.updateTexture()
	}
}

func (w *Container) AdoptChild(c ElementI) {
	w.BaseElement.AdoptChild(c)
	w.reflow()
}

func (w *Container) reflow() {
	var y int32
	if w.Style.Display.Has(COLUMNS) {
		if w.Style.Direction.Has(REVERSE) {
			y := w.h
			for i := len(w.Children) - 1; i >= 0; i-- {
				child := w.Children[i]
				switch c := child.(type) {
				case *Container:
					c.reflow()
				}
				y -= child.GetHeight()
				child.SetY(y)
			}
		} else {
			for _, child := range w.Children {
				switch c := child.(type) {
				case *Container:
					c.reflow()
				}
				child.SetY(y)
				y += child.GetHeight()
			}
		}
	}
}

// Destroy the window, clearing the SDL context and destroying the SDLWindow if it is a top-level window.
func (w *Container) Destroy() {
	if w.SDLTexture != nil {
		w.SDLTexture.Destroy()
	}

	w.BaseElement.Destroy()
}
