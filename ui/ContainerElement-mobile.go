// +build mobile

package ui

import (
	"golang.org/x/mobile/gl"
)

// Container is a UI element that represents a texture-backed containing element.
type Container struct {
	BaseElement
	GLTexture gl.Texture

	ContainerRenderFunc ContainerRenderFunc
}

func (w *Container) updateTexture() (err error) {
	if w.Parent == nil {
		return
	}
	//
	if !w.Context.GLContext.IsTexture(w.GLTexture) {

	}

	return
}

// Render the window, its renderer function, and its children to its texture,
// thereafter rendering its texture to a Parent if it exists or to the screen
// if it is a top-level window.
func (w *Container) Render() {
	if w.IsHidden() {
		return
	}
	//w.BaseElement.Render()
}

// CalculateStyle recalculates the style and updates the Container texture if it is dirty. See BaseElement.CalculateStyle().
func (w *Container) CalculateStyle() {
	w.BaseElement.CalculateStyle()
	if w.IsDirty() {
		w.updateTexture()
	}
}

// Destroy the window, clearing the SDL context and destroying the SDLWindow if it is a top-level window.
func (w *Container) Destroy() {
	w.Parent.DisownChild(w)

	if w.Context.GLContext.IsTexture(w.GLTexture) {
		w.Context.GLContext.DeleteTexture(w.GLTexture)
	}

	for _, child := range w.Children {
		child.Destroy()
	}
}
