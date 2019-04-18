// +build !MOBILE
package UI

import (
	"github.com/veandco/go-sdl2/sdl"
	"strings"
)

type InputElement struct {
	BaseElement
	SDL_texture *sdl.Texture
	Image       []byte
	tw          int32 // Texture width
	th          int32 // Texture height
	cursor      int
	composition []rune
	isPassword  bool
	placeholder string
}

type InputElementConfig struct {
	Style       Style
	Value       string
	Events      Events
	Password    bool
	Placeholder string
}

func NewInputElement(c InputElementConfig) ElementI {
	i := InputElement{}
	i.This = ElementI(&i)
	i.Style.Set(c.Style)
	i.composition = []rune(c.Value)
	i.cursor = len(i.composition)
	i.SyncComposition()
	i.Events = c.Events
	i.isPassword = c.Password
	i.placeholder = c.Placeholder
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
		Y: ty,
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
			X: tx + int32(cursor_start) - 1,
			Y: ty,
			W: 1,
			H: int32(cursor_height),
		}
		t.Context.Renderer.FillRect(&cursor_dst)
	}
	t.BaseElement.Render()
}

func (t *InputElement) SetValue(value string) (err error) {
	t.Value = value
	var render_str string
	render_color := sdl.Color{
		t.Style.ForegroundColor.R,
		t.Style.ForegroundColor.G,
		t.Style.ForegroundColor.B,
		t.Style.ForegroundColor.A,
	}
	if t.Context == nil || t.Context.Font == nil {
		return
	}
	if t.SDL_texture != nil {
		t.SDL_texture.Destroy()
		t.SDL_texture = nil
	}

	if len(value) == 0 {
		if len(t.placeholder) == 0 {
			render_str = " "
		} else {
			render_str = t.placeholder
			render_color.A = render_color.A / 2
		}
	} else {
		if t.isPassword {
			render_str = strings.Repeat("*", len(value))
		} else {
			render_str = value
		}
	}

	surface, err := t.Context.Font.RenderUTF8Blended(render_str, render_color)
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

func (t *InputElement) CalculateStyle() {
	if t.SDL_texture == nil {
		t.SetValue(t.Value)
	}
	t.BaseElement.CalculateStyle()
}

func (i *InputElement) OnFocus() bool {
	sdl.StartTextInput()
	return i.BaseElement.OnFocus()
}
func (i *InputElement) OnBlur() bool {
	sdl.StopTextInput()
	return i.BaseElement.OnBlur()
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
