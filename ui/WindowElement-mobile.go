// +build mobile

package ui

import (
	"golang.org/x/mobile/gl"
)

// Window is a UI element that can either represent a standalone OS window or a
// sub-window contained within another Window.
type Window struct {
	BaseElement

	GLTexture gl.Texture

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

	if err != nil {
		return err
	}

	w.CalculateStyle()

	if err != nil {
		return err
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

	//w.Context.GLContext.BindFramebuffer(gl.FRAMEBUFFER, gl.Framebuffer{0})

	w.Context.GLContext.ClearColor(1, 0, 0, 1)
	w.Context.GLContext.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
	w.Context.GLContext.Enable(gl.BLEND)
	w.Context.GLContext.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	w.Context.GLContext.Enable(gl.CULL_FACE)
	w.Context.GLContext.Disable(gl.DEPTH_TEST)

	if w.RenderFunc != nil {
		w.RenderFunc(w)
	}
	if w.Style.BackgroundColor.A > 0 {
		/*w.Context.Renderer.SetFillColor(nanovgo.RGBA(
			w.Style.BackgroundColor.R,
			w.Style.BackgroundColor.G,
			w.Style.BackgroundColor.B,
			w.Style.BackgroundColor.A,
		))

		w.Context.Renderer.BeginPath()

		w.Context.Renderer.Rect(float32(w.x), float32(w.y), float32(w.w), float32(w.h))

		w.Context.Renderer.ClosePath()
		w.Context.Renderer.Fill()*/
	}
	w.BaseElement.Render()

	//w.Context.GLContext.BindFramebuffer(gl.FRAMEBUFFER, gl.Framebuffer{0})
	w.Context.GLContext.Flush()
}

// Destroy the window, clearing the SDL context and destroying the SDLWindow if it is a top-level window.
func (w *Window) Destroy() {
	for _, child := range w.Children {
		child.Destroy()
	}

	if w.Context.GLContext.IsTexture(w.GLTexture) {
		w.Context.GLContext.DeleteTexture(w.GLTexture)
	}
}
