// +build mobile

package ui

import (
	"golang.org/x/mobile/gl"
)

// ButtonElement is the element type responsible for receiving mouse or touch
// events and updating its rendering appropriately.
type ButtonElement struct {
	BaseElement

	GLTexture gl.Texture
	tw        int32 // Texture width
	th        int32 // Texture height
}

// Destroy destroys the underlying SDL texture used for text rendering.
func (b *ButtonElement) Destroy() {
	if b.GLTexture.Value > 0 {
		b.Context.GLContext.DeleteTexture(b.GLTexture)
	}
}

// Render draws the button and its state using the element's renderer contexb.
func (b *ButtonElement) Render() {
	b.BaseElement.Render()
}

// SetValue sets the text value of the button and updates the texture as
// needed.
func (b *ButtonElement) SetValue(value string) (err error) {
	b.Value = value
	b.Dirty = true
	return
}

// CalculateStyle creates the SDL texture if it doesn't exist before calling
// BaseElement.CalculateStyle()
func (b *ButtonElement) CalculateStyle() {
	if b.GLTexture.Value == 0 {
		b.SetValue(b.Value)
	}
	b.BaseElement.CalculateStyle()
}
