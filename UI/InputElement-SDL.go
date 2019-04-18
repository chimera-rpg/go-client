// +build !MOBILE
package UI

import (
	"github.com/veandco/go-sdl2/sdl"
)

type InputElement struct {
	BaseElement
	SDL_texture *sdl.Texture
	Image       []byte
	tw          int32 // Texture width
	th          int32 // Texture height
	cursor      int
	composition []rune
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
	i.composition = []rune(c.Value)
	i.cursor = len(i.composition)
	i.SyncComposition()
	i.Events = c.Events
	i.Focusable = true

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
	if t.Focused {
		// Draw our border
		if t.Style.BackgroundColor.A > 0 {
			dst := sdl.Rect{
				X: t.x,
				Y: t.y,
				W: t.w,
				H: t.h,
			}
			t.Context.Renderer.SetDrawColor(255-t.Style.BackgroundColor.R, 255-t.Style.BackgroundColor.G, 255-t.Style.BackgroundColor.B, 255-t.Style.BackgroundColor.A)
			t.Context.Renderer.DrawRect(&dst)
		}
		// Get and draw our cursor position
		cursor_start, cursor_height, _ := t.Context.Font.SizeUTF8(string(t.composition[:t.cursor]))
		t.Context.Renderer.SetDrawColor(t.Style.ForegroundColor.R, t.Style.ForegroundColor.G, t.Style.ForegroundColor.B, t.Style.ForegroundColor.A)
		cursor_dst := sdl.Rect{
			X: t.x + int32(cursor_start) + t.pl - 1,
			Y: t.y + t.pt,
			W: 1,
			H: int32(cursor_height),
		}
		t.Context.Renderer.FillRect(&cursor_dst)
	}
	t.BaseElement.Render()
}

func (t *InputElement) SetValue(value string) (err error) {
	t.Value = value
	if len(t.Value) == 0 {
		t.Value = " "
	}
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

func (i *InputElement) OnFocus() bool {
	sdl.StartTextInput()
	i.Dirty = true
	if i.Events.OnFocus != nil {
		return i.Events.OnFocus()
	}
	return true
}
func (i *InputElement) OnBlur() bool {
	sdl.StopTextInput()
	i.Dirty = true
	if i.Events.OnBlur != nil {
		return i.Events.OnBlur()
	}
	return true
}

func (i *InputElement) SyncComposition() {
	i.SetValue(string(i.composition))
}

func (i *InputElement) OnKeyDown(key uint8, modifiers uint16) bool {
	switch key {
	case 27: // esc
		//BlurFocusedElement()
	case 8: // backspace
		if i.cursor > 0 {
			i.composition = append(i.composition[:i.cursor-1], i.composition[i.cursor:]...)
			i.cursor -= 1
		}
	case 127: // delete
		if i.cursor < len(i.composition) {
			i.composition = append(i.composition[:i.cursor], i.composition[i.cursor+1:]...)
		}
	case 9: // tab
	case 79: // right
		i.cursor += 1
		if i.cursor > len(i.composition) {
			i.cursor = len(i.composition)
		}
	case 80: // left
		i.cursor -= 1
		if i.cursor < 0 {
			i.cursor = 0
		}
	case 81: // down
		i.cursor = 0
	case 82: // up
		i.cursor = len(i.composition)
	}
	i.SyncComposition()
	if i.Events.OnKeyDown != nil {
		return i.Events.OnKeyDown(key, modifiers)
	}
	return true
}

func (i *InputElement) OnKeyUp(key uint8, modifiers uint16) bool {
	if i.Events.OnKeyUp != nil {
		return i.Events.OnKeyUp(key, modifiers)
	}
	return true
}

func (i *InputElement) OnTextInput(str string) bool {
	runes := []rune(str)
	i.composition = append(i.composition[:i.cursor], append(runes, i.composition[i.cursor:]...)...)
	i.cursor += len(runes)
	i.SyncComposition()
	if i.Events.OnTextInput != nil {
		return i.Events.OnTextInput(str)
	}
	return true
}
func (i *InputElement) OnTextEdit(str string, start int32, length int32) bool {
	if i.Events.OnTextEdit != nil {
		return i.Events.OnTextEdit(str, start, length)
	}
	return true
}
