//go:build mobile
// +build mobile

package ui

import (
	"golang.org/x/mobile/gl"
)

// InputElement is the element that handles user input and display within a
// field.
type InputElement struct {
	BaseElement
	GLTexture gl.Texture

	tw            int32 // Texture width
	th            int32 // Texture height
	cursor        int
	composition   []rune
	isPassword    bool
	placeholder   string
	submitOnEnter bool
	clearOnSubmit bool
	blurOnSubmit  bool
}

// Destroy cleans up the InputElement's resources.
func (i *InputElement) Destroy() {
	if i.Context.GLContext.IsTexture(i.GLTexture) {
		i.Context.GLContext.DeleteTexture(i.GLTexture)
	}
}

// Render renders the InputElement to the rendering context, with various
// conditionally rendered aspects to represent state.
func (i *InputElement) Render() {
	if i.IsHidden() {
		return
	}
	if !i.Context.GLContext.IsTexture(i.GLTexture) {
		i.SetValue(i.Value)
	}

	i.BaseElement.Render()
}

// SetValue sets the text value of the input field and recreates and renders
// to its underlying texture.
func (i *InputElement) SetValue(value string) (err error) {
	i.Value = value

	i.Dirty = true
	i.OnChange()
	return
}

// CalculateStyle sets the SDLTexture if it doesn't exist before calculating
// the style.
func (i *InputElement) CalculateStyle() {
	if !i.Context.GLContext.IsTexture(i.GLTexture) {
		i.SetValue(i.Value)
	}
	i.BaseElement.CalculateStyle()
}

// OnFocus calls sdl.StartTextInput
func (i *InputElement) OnFocus() bool {
	//sdl.StartTextInput()
	return i.BaseElement.OnFocus()
}

// OnBlur calls sdl.StopTextInput
func (i *InputElement) OnBlur() bool {
	//sdl.StopTextInput()
	return i.BaseElement.OnBlur()
}
