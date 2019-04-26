// +build !MOBILE

package ui

import (
	"github.com/veandco/go-sdl2/sdl"
)

// TextElement is our main element for handling and drawing text.
type TextElement struct {
	BaseElement
	SDLTexture *sdl.Texture
	tw         int32 // Texture width
	th         int32 // Texture height
}

// TextElementConfig is the configuration object passed to NewTextElement.
type TextElementConfig struct {
	Style  string
	Value  string
	Events Events
}

// TextElementStyle is our default styling for TextElements.
var TextElementStyle = `
	ForegroundColor 0 0 0 255
	Padding 6
	MinH 12
	H 7%
	MaxH 30
`

// NewTextElement creates a new TextElement from the passed configuration.
func NewTextElement(c TextElementConfig) ElementI {
	t := TextElement{}
	t.This = ElementI(&t)
	t.Style.Parse(TextElementStyle)
	t.Style.Parse(c.Style)
	t.SetValue(c.Value)
	t.Events = c.Events

	t.OnCreated()

	return ElementI(&t)
}

// Destroy handles the destruction of the underlying texture.
func (t *TextElement) Destroy() {
	if t.SDLTexture != nil {
		t.SDLTexture.Destroy()
	}
}

// Render renders our base styling before rendering its text texture using
// the context renderer.
func (t *TextElement) Render() {
	if t.IsHidden() {
		return
	}
	if t.SDLTexture == nil {
		t.SetValue(t.Value)
	}
	if t.Style.BackgroundColor.A > 0 {
		dst := sdl.Rect{
			X: t.x,
			Y: t.y,
			W: t.w,
			H: t.h,
		}
		t.Context.Renderer.SetDrawColor(t.Style.BackgroundColor.R, t.Style.BackgroundColor.G, t.Style.BackgroundColor.B, t.Style.BackgroundColor.A)
		t.Context.Renderer.FillRect(&dst)
	}
	// Render text
	tx := t.x + t.pl
	ty := t.y + t.pt
	if t.Style.ContentOrigin.Has(CENTERX) {
		tx += t.w/2 - t.tw/2 - t.pr
	}
	if t.Style.ContentOrigin.Has(CENTERY) {
		ty += t.h/2 - t.th/2 - t.pb
	}
	dst := sdl.Rect{
		X: tx,
		Y: ty,
		W: t.tw,
		H: t.th,
	}
	t.Context.Renderer.Copy(t.SDLTexture, nil, &dst)
	t.BaseElement.Render()
}

// SetValue sets the text value for the TextElement, (re)creating the
// underlying SDL texture as needed.
func (t *TextElement) SetValue(value string) (err error) {
	t.Value = value
	if t.Context == nil || t.Context.Font == nil {
		return
	}
	if t.SDLTexture != nil {
		t.SDLTexture.Destroy()
		t.SDLTexture = nil
	}
	surface, err := t.Context.Font.RenderUTF8Blended(t.Value,
		sdl.Color{
			R: t.Style.ForegroundColor.R,
			G: t.Style.ForegroundColor.G,
			B: t.Style.ForegroundColor.B,
			A: t.Style.ForegroundColor.A,
		})
	defer surface.Free()
	if err != nil {
		panic(err)
	}
	t.SDLTexture, err = t.Context.Renderer.CreateTextureFromSurface(surface)
	if err != nil {
		panic(err)
	}

	t.tw = surface.W
	t.th = surface.H
	if t.Style.ResizeToContent {
		t.Style.W.Set(float64(surface.W))
		t.Style.H.Set(float64(surface.H))
	}
	t.Dirty = true
	return
}

// CalculateStyle is the same as BaseElement with the addition of always
// creating the SDL texture if it has not been created.
func (t *TextElement) CalculateStyle() {
	if t.SDLTexture == nil {
		t.SetValue(t.Value)
	}
	t.BaseElement.CalculateStyle()
}
