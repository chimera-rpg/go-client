// +build !MOBILE
package UI

import (
	"github.com/veandco/go-sdl2/sdl"
)

type TextElement struct {
	BaseElement
	SDL_texture *sdl.Texture
	tw          int32 // Texture width
	th          int32 // Texture height
}

type TextElementConfig struct {
	Style  string
	Value  string
	Events Events
}

var TextElementStyle = `
	ForegroundColor 0 0 0 255
`

func NewTextElement(c TextElementConfig) ElementI {
	t := TextElement{}
	t.This = ElementI(&t)
	t.Style.Parse(TextElementStyle)
	t.Style.Parse(c.Style)
	t.SetValue(c.Value)
	t.Events = c.Events

	return ElementI(&t)
}

func (t *TextElement) Destroy() {
	if t.SDL_texture != nil {
		t.SDL_texture.Destroy()
	}
}

func (t *TextElement) Render() {
	if t.IsHidden() {
		return
	}
	if t.SDL_texture == nil {
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
	dst := sdl.Rect{
		X: t.x + t.pl,
		Y: t.y + t.pt,
		W: t.tw,
		H: t.th,
	}
	t.Context.Renderer.Copy(t.SDL_texture, nil, &dst)
	t.BaseElement.Render()
}

func (t *TextElement) SetValue(value string) (err error) {
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
	t.Style.W.Set(float64(surface.W))
	t.Style.H.Set(float64(surface.H))
	t.Dirty = true
	return
}

func (t *TextElement) CalculateStyle() {
	if t.SDL_texture == nil {
		t.SetValue(t.Value)
	}
	t.BaseElement.CalculateStyle()
}
