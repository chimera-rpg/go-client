// +build !MOBILE

package ui

import (
	"github.com/veandco/go-sdl2/sdl"
)

// ButtonElement is the element type responsible for receiving mouse or touch
// events and updating its rendering appropriately.
type ButtonElement struct {
	BaseElement
	SDLTexture *sdl.Texture
	tw         int32 // Texture width
	th         int32 // Texture height
}

// ButtonElementConfig provides the configuration for a new ButtonElemenb.
type ButtonElementConfig struct {
	Style  string
	Value  string
	Events Events
}

// ButtonElementStyle is the default style for ButtonElements.
var ButtonElementStyle = `
	ForegroundColor 255 255 255 255
	BackgroundColor 139 139 186 128
	ContentOrigin CenterX CenterY
	MinH 12
	H 7%
	MaxH 40
`

// NewButtonElement creates a ButtonElement using the passed configuration.
func NewButtonElement(c ButtonElementConfig) ElementI {
	b := ButtonElement{}
	b.This = ElementI(&b)
	b.Holdable = true
	b.Focusable = true
	b.Style.Parse(ButtonElementStyle)
	b.Style.Parse(c.Style)
	b.SetValue(c.Value)
	b.Events = c.Events
	b.OnCreated()

	return ElementI(&b)
}

// Destroy destroys the underlying SDL texture used for text rendering.
func (b *ButtonElement) Destroy() {
	if b.SDLTexture != nil {
		b.SDLTexture.Destroy()
	}
}

// Render draws the button and its state using the element's renderer contexb.
func (b *ButtonElement) Render() {
	if b.IsHidden() {
		return
	}
	if b.SDLTexture == nil {
		b.SetValue(b.Value)
	}
	heldOffset := int32(0)
	if b.Style.BackgroundColor.A > 0 {
		offsetY := int32(b.h / 10)
		if b.Held {
			heldOffset = offsetY
		}
		// Draw top portion
		dst := sdl.Rect{
			X: b.x,
			Y: b.y + heldOffset,
			W: b.w,
			H: b.h - offsetY,
		}
		b.Context.Renderer.SetDrawColor(b.Style.BackgroundColor.R, b.Style.BackgroundColor.G, b.Style.BackgroundColor.B, b.Style.BackgroundColor.A)
		b.Context.Renderer.FillRect(&dst)
		if !b.Held {
			// Draw bottom portion
			dst = sdl.Rect{
				X: b.x,
				Y: b.y + (b.h - offsetY),
				W: b.w,
				H: offsetY,
			}
			b.Context.Renderer.SetDrawColor(b.Style.BackgroundColor.R-64, b.Style.BackgroundColor.G-64, b.Style.BackgroundColor.B-64, b.Style.BackgroundColor.A)
			b.Context.Renderer.FillRect(&dst)
		}
		if b.Focused {
			// Draw our border
			dst := sdl.Rect{
				X: b.x,
				Y: b.y + heldOffset,
				W: b.w,
				H: b.h - heldOffset,
			}
			b.Context.Renderer.SetDrawColor(255-b.Style.BackgroundColor.R, 255-b.Style.BackgroundColor.G, 255-b.Style.BackgroundColor.B, 255-b.Style.BackgroundColor.A)
			b.Context.Renderer.DrawRect(&dst)
		}
	}
	// Render text texture
	tx := b.x + b.pl
	ty := b.y + b.pt
	if b.Style.ContentOrigin.Has(CENTERX) {
		tx += b.w/2 - b.tw/2 - b.pr
	}
	if b.Style.ContentOrigin.Has(CENTERY) {
		ty += b.h/2 - b.th/2 - b.pb
	}
	dst := sdl.Rect{
		X: tx,
		Y: ty + heldOffset,
		W: b.tw,
		H: b.th,
	}
	b.Context.Renderer.Copy(b.SDLTexture, nil, &dst)
	b.BaseElement.Render()
}

// SetValue sets the text value of the button and updates the SDL texture as
// needed.
func (b *ButtonElement) SetValue(value string) (err error) {
	b.Value = value
	if b.Context == nil || b.Context.Font == nil {
		return
	}
	if b.SDLTexture != nil {
		b.SDLTexture.Destroy()
		b.SDLTexture = nil
	}
	surface, err := b.Context.Font.RenderUTF8Blended(b.Value,
		sdl.Color{
			R: b.Style.ForegroundColor.R,
			G: b.Style.ForegroundColor.G,
			B: b.Style.ForegroundColor.B,
			A: b.Style.ForegroundColor.A,
		})
	defer surface.Free()
	if err != nil {
		panic(err)
	}
	b.SDLTexture, err = b.Context.Renderer.CreateTextureFromSurface(surface)
	if err != nil {
		panic(err)
	}

	b.tw = surface.W
	b.th = surface.H
	if b.Style.ResizeToContent {
		b.Style.W.Set(float64(surface.W))
		b.Style.H.Set(float64(surface.H))
	}
	b.Dirty = true
	return
}

// CalculateStyle creates the SDL texture if it doesn't exist before calling
// BaseElement.CalculateStyle()
func (b *ButtonElement) CalculateStyle() {
	if b.SDLTexture == nil {
		b.SetValue(b.Value)
	}
	b.BaseElement.CalculateStyle()
}

// OnKeyDown sets the button's held state when the enter key is pressed.
func (b *ButtonElement) OnKeyDown(key uint8, modifiers uint16) bool {
	switch key {
	case 13: // Activate button when enter is hit
		b.SetHeld(true)
	}
	return false
}

// OnKeyUp unsets the button's held state and triggers the OnMouseButtonUp
// method when the enter key is released.
func (b *ButtonElement) OnKeyUp(key uint8, modifiers uint16) bool {
	switch key {
	case 13: // Activate button when enter is released
		b.SetHeld(false)
		b.OnMouseButtonUp(1, 0, 0)
	}
	return false
}
