// +build mobile

package ui

import (
	"golang.org/x/mobile/gl"
)

// ImageElement is the element responsible for rendering an image.
type ImageElement struct {
	BaseElement
	GLTexture gl.Texture

	Image []byte
	tw    int32 // Texture width
	th    int32 // Texture height
}

// Destroy destroys the underlying ImageElement.
func (i *ImageElement) Destroy() {
	if i.Context.GLContext.IsTexture(i.GLTexture) {
		i.Context.GLContext.DeleteTexture(i.GLTexture)
	}
}

// Render renders the ImageElement to the screen.
func (i *ImageElement) Render() {
	if i.IsHidden() {
		return
	}
	if i.GLTexture.Value == 0 {
		i.SetImage(i.Image)
	}

	i.BaseElement.Render()
}

// SetImage sets the underlying texture to the passed PNG byte slice.
func (i *ImageElement) SetImage(png []byte) {
	if i.Context == nil {
		return
	}

	i.Dirty = true
}
