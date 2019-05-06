// +build mobile

package ui

import (
	"golang.org/x/mobile/gl"
)

// TextElement is our main element for handling and drawing text.
type TextElement struct {
	BaseElement
	GLTexture gl.Texture
	tw        int32 // Texture width
	th        int32 // Texture height
}

// Destroy handles the destruction of the underlying texture.
func (t *TextElement) Destroy() {
	if t.Context.GLContext.IsTexture(t.GLTexture) {
		t.Context.GLContext.DeleteTexture(t.GLTexture)
	}
}

// Render renders our base styling before rendering its text texture using
// the context renderer.
func (t *TextElement) Render() {
	if t.IsHidden() {
		return
	}
	if !t.Context.GLContext.IsTexture(t.GLTexture) {
		t.SetValue(t.Value)
	}
	t.BaseElement.Render()
}

// SetValue sets the text value for the TextElement, (re)creating the
// underlying SDL texture as needed.
func (t *TextElement) SetValue(value string) (err error) {
	t.Value = value
	if value == "" {
		value = " "
	}

	t.Dirty = true
	t.OnChange()
	return
}

// CalculateStyle is the same as BaseElement with the addition of always
// creating the SDL texture if it has not been created.
func (t *TextElement) CalculateStyle() {
	if !t.Context.GLContext.IsTexture(t.GLTexture) {
		t.SetValue(t.Value)
	}
	t.BaseElement.CalculateStyle()
}
