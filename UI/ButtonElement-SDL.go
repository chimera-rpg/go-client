// +build !MOBILE
package UI

import (
	"github.com/veandco/go-sdl2/sdl"
)

type ButtonElement struct {
	BaseElement
	SDL_texture *sdl.Texture
	tw          int32 // Texture width
	th          int32 // Texture height
}

type ButtonElementConfig struct {
	Style  Style
	Value  string
	Events Events
}

func NewButtonElement(c ButtonElementConfig) ElementI {
	t := ButtonElement{}
	t.This = ElementI(&t)
	t.Holdable = true
	t.Focusable = true
	t.Style.Set(c.Style)
	t.SetValue(c.Value)
	t.Events = c.Events

	return ElementI(&t)
}

func (t *ButtonElement) Destroy() {
	if t.SDL_texture != nil {
		t.SDL_texture.Destroy()
	}
}

func (t *ButtonElement) Render() {
	if t.IsHidden() {
		return
	}
	if t.SDL_texture == nil {
		t.SetValue(t.Value)
	}
	held_offset := int32(0)
	if t.Style.BackgroundColor.A > 0 {
		offset_y := int32(t.h / 10)
		if t.Held {
			held_offset = offset_y
		}
		// Draw top portion
		dst := sdl.Rect{
			X: t.x,
			Y: t.y + held_offset,
			W: t.w,
			H: t.h - offset_y,
		}
		t.Context.Renderer.SetDrawColor(t.Style.BackgroundColor.R, t.Style.BackgroundColor.G, t.Style.BackgroundColor.B, t.Style.BackgroundColor.A)
		t.Context.Renderer.FillRect(&dst)
		if !t.Held {
			// Draw bottom portion
			dst = sdl.Rect{
				X: t.x,
				Y: t.y + (t.h - offset_y),
				W: t.w,
				H: offset_y,
			}
			t.Context.Renderer.SetDrawColor(t.Style.BackgroundColor.R-64, t.Style.BackgroundColor.G-64, t.Style.BackgroundColor.B-64, t.Style.BackgroundColor.A)
			t.Context.Renderer.FillRect(&dst)
		}
		if t.Focused {
			// Draw our border
			dst := sdl.Rect{
				X: t.x,
				Y: t.y + held_offset,
				W: t.w,
				H: t.h - held_offset,
			}
			t.Context.Renderer.SetDrawColor(255-t.Style.BackgroundColor.R, 255-t.Style.BackgroundColor.G, 255-t.Style.BackgroundColor.B, 255-t.Style.BackgroundColor.A)
			t.Context.Renderer.DrawRect(&dst)
		}
	}
	// Render text texture
	tx := t.x + t.pl
	ty := t.y + t.pt
	if (t.Style.CenterContent & CENTERX) == CENTERX {
		tx += t.w/2 - t.tw/2
	}
	if (t.Style.CenterContent & CENTERY) == CENTERY {
		ty += t.h/2 - t.th/2
	}
	dst := sdl.Rect{
		X: tx,
		Y: ty + held_offset,
		W: t.tw,
		H: t.th,
	}
	t.Context.Renderer.Copy(t.SDL_texture, nil, &dst)
	t.BaseElement.Render()
}

func (t *ButtonElement) SetValue(value string) (err error) {
	t.Value = value
	if t.Context == nil || t.Context.Font == nil {
		return
	}
	if t.SDL_texture != nil {
		t.SDL_texture.Destroy()
		t.SDL_texture = nil
	}
	surface, err := t.Context.Font.RenderUTF8Blended(t.Value,
		sdl.Color{
			t.Style.ForegroundColor.R,
			t.Style.ForegroundColor.G,
			t.Style.ForegroundColor.B,
			t.Style.ForegroundColor.A,
		})
	defer surface.Free()
	if err != nil {
		panic(err)
	}
	t.SDL_texture, err = t.Context.Renderer.CreateTextureFromSurface(surface)
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

func (t *ButtonElement) CalculateStyle() {
	if t.SDL_texture == nil {
		t.SetValue(t.Value)
	}
	t.BaseElement.CalculateStyle()
}

func (b *ButtonElement) OnKeyDown(key uint8, modifiers uint16) bool {
	switch key {
	case 13: // Activate button when enter is hit
		b.SetHeld(true)
	}
	return false
}
func (b *ButtonElement) OnKeyUp(key uint8, modifiers uint16) bool {
	switch key {
	case 13: // Activate button when enter is released
		b.SetHeld(false)
		b.OnMouseButtonUp(1, 0, 0)
	}
	return false
}
