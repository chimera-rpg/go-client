// +build !MOBILE
package UI

import (
	"github.com/veandco/go-sdl2/sdl"
)

type InputElement struct {
	BaseElement
	SDL_texture      *sdl.Texture
	Image            []byte
	tw               int32 // Texture width
	th               int32 // Texture height
	ChildTextElement ElementI
}

type InputElementConfig struct {
	Style  Style
	Value  string
	Events Events
}

func NewInputElement(c InputElementConfig) ElementI {
	i := InputElement{}
	i.This = ElementI(&i)
	i.Style.Set(c.Style)
	i.SetValue(c.Value)
	i.Events = c.Events
	i.ChildTextElement = NewTextElement(TextElementConfig{
		Style: Style{
			PaddingLeft: Number{
				Percentage: true,
				Value:      5,
			},
			PaddingRight: Number{
				Percentage: true,
				Value:      5,
			},
			PaddingTop: Number{
				Percentage: true,
				Value:      5,
			},
			PaddingBottom: Number{
				Percentage: true,
				Value:      5,
			},
			X: Number{
				Value: 0,
			},
			Y: Number{
				Percentage: true,
				Value:      50,
			},
		},
		Value: "Dummy text",
	})

	return ElementI(&i)
}

func (t *InputElement) Destroy() {
	if t.SDL_texture != nil {
		t.SDL_texture.Destroy()
	}
}

func (t *InputElement) Render() {
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

func (t *InputElement) SetValue(value string) (err error) {
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

func (t *InputElement) CalculateStyle() {
	if t.SDL_texture == nil {
		t.SetValue(t.Value)
	}
	t.BaseElement.CalculateStyle()
}

func (i *InputElement) OnAdopted(parent ElementI) {
	i.BaseElement.OnAdopted(parent)
	i.AdoptChild(i.ChildTextElement)
}
