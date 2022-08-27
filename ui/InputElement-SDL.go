//go:build !mobile
// +build !mobile

package ui

import (
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

// InputElement is the element that handles user input and display within a
// field.
type InputElement struct {
	BaseElement
	SDLTexture    *sdl.Texture
	Image         []byte
	tw            int32 // Texture width
	th            int32 // Texture height
	cursor        int
	composition   []rune
	isPassword    bool
	placeholder   string
	submitOnEnter bool
	clearOnSubmit bool
	blurOnSubmit  bool
	keysHeld      map[uint8]bool
}

// Destroy cleans up the InputElement's resources.
func (i *InputElement) Destroy() {
	if i.SDLTexture != nil {
		i.SDLTexture.Destroy()
	}
	i.BaseElement.Destroy()
}

// Render renders the InputElement to the rendering context, with various
// conditionally rendered aspects to represent state.
func (i *InputElement) Render() {
	if i.IsHidden() {
		return
	}
	if i.SDLTexture == nil {
		i.SetValue(i.Value)
	}
	if i.Style.BackgroundColor.A > 0 {
		dst := sdl.Rect{
			X: i.x,
			Y: i.y,
			W: i.w,
			H: i.h,
		}
		i.Context.Renderer.SetDrawColor(i.Style.BackgroundColor.R, i.Style.BackgroundColor.G, i.Style.BackgroundColor.B, i.Style.BackgroundColor.A)
		i.Context.Renderer.FillRect(&dst)
	}
	// Render text texture
	tx := i.x + i.pl
	ty := i.y + i.pt
	if i.Style.ContentOrigin.Has(CENTERX) {
		tx += i.w/2 - i.tw/2 - i.pr
	}
	if i.Style.ContentOrigin.Has(CENTERY) {
		ty += i.h/2 - i.th/2 - i.pb
	}
	dst := sdl.Rect{
		X: tx,
		Y: ty,
		W: i.tw,
		H: i.th,
	}
	i.Context.Renderer.Copy(i.SDLTexture, nil, &dst)
	if i.Focused {
		// Draw our border
		if i.Style.BackgroundColor.A > 0 {
			dst := sdl.Rect{
				X: i.x,
				Y: i.y,
				W: i.w,
				H: i.h,
			}
			i.Context.Renderer.SetDrawColor(255-i.Style.BackgroundColor.R, 255-i.Style.BackgroundColor.G, 255-i.Style.BackgroundColor.B, 255-i.Style.BackgroundColor.A)
			i.Context.Renderer.DrawRect(&dst)
		}
		// Get and draw our cursor position
		cursorStart, cursorHeight, _ := i.Context.Font.SizeUTF8(string(i.composition[:i.cursor]))
		i.Context.Renderer.SetDrawColor(i.Style.ForegroundColor.R, i.Style.ForegroundColor.G, i.Style.ForegroundColor.B, i.Style.ForegroundColor.A)
		cursorDst := sdl.Rect{
			X: tx + int32(cursorStart) - 1,
			Y: ty,
			W: 1,
			H: int32(cursorHeight),
		}
		i.Context.Renderer.FillRect(&cursorDst)
	}
	i.BaseElement.Render()
}

// SetValue sets the text value of the input field and recreates and renders
// to its underlying texture.
func (i *InputElement) SetValue(value string) (err error) {
	i.Value = value
	var renderStr string
	renderColor := sdl.Color{
		R: i.Style.ForegroundColor.R,
		G: i.Style.ForegroundColor.G,
		B: i.Style.ForegroundColor.B,
		A: i.Style.ForegroundColor.A,
	}
	if i.Context == nil || i.Context.Font == nil {
		return
	}
	if i.SDLTexture != nil {
		i.SDLTexture.Destroy()
		i.SDLTexture = nil
	}

	if len(value) == 0 {
		// NOTE: RenderUTF8Blended cannot take a zero-length string, so we're
		// populating a blank space if needed.
		if len(i.placeholder) == 0 {
			renderStr = " "
		} else {
			renderStr = i.placeholder
			renderColor.A = renderColor.A / 2
		}
	} else {
		if i.isPassword {
			renderStr = strings.Repeat("*", len(value))
		} else {
			renderStr = value
		}
	}

	surface, err := i.Context.Font.RenderUTF8Blended(renderStr, renderColor)
	defer surface.Free()
	if err != nil {
		panic(err)
	}
	i.SDLTexture, err = i.Context.Renderer.CreateTextureFromSurface(surface)
	if err != nil {
		panic(err)
	}

	i.tw = surface.W
	i.th = surface.H
	if i.Style.Resize.Has(TOCONTENT) {
		i.Style.W.Set(float64(surface.W))
		i.Style.H.Set(float64(surface.H))
	}
	i.Dirty = true
	i.OnChange()
	return
}

// CalculateStyle sets the SDLTexture if it doesn't exist before calculating
// the style.
func (i *InputElement) CalculateStyle() {
	if i.SDLTexture == nil {
		i.SetValue(i.Value)
	}
	i.BaseElement.CalculateStyle()
}

// OnFocus calls sdl.StartTextInput
func (i *InputElement) OnFocus() bool {
	sdl.StartTextInput()
	i.keysHeld = make(map[uint8]bool)
	return i.BaseElement.OnFocus()
}

// OnBlur calls sdl.StopTextInput
func (i *InputElement) OnBlur() bool {
	sdl.StopTextInput()
	i.keysHeld = make(map[uint8]bool)
	return i.BaseElement.OnBlur()
}
