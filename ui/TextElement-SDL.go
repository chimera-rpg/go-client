//go:build !mobile
// +build !mobile

package ui

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

// TextElement is our main element for handling and drawing text.
type TextElement struct {
	BaseElement
	SDLTexture *sdl.Texture
	tw         int32 // Texture width
	th         int32 // Texture height
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
	if t.Style.Origin.Has(BOTTOM) {
		//ty -= t.h
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
	if value == "" {
		value = " "
	}
	if t.Context == nil || t.Context.Font == nil {
		return
	}
	if t.SDLTexture != nil {
		t.SDLTexture.Destroy()
		t.SDLTexture = nil
	}
	// Create text Outline
	var textSurface, outlineSurface *sdl.Surface
	var textTexture, outlineTexture *sdl.Texture

	if t.Style.OutlineColor.A > 0 {
		outlineSurface, err = t.Context.OutlineFont.RenderUTF8Blended(value,
			sdl.Color{
				R: t.Style.OutlineColor.R,
				G: t.Style.OutlineColor.G,
				B: t.Style.OutlineColor.B,
				A: 255,
			},
		)
		if err != nil {
			panic(err)
		}
		defer outlineSurface.Free()
		if outlineTexture, err = t.Context.Renderer.CreateTextureFromSurface(outlineSurface); err != nil {
			panic(err)
		}
		defer outlineTexture.Destroy()
	}
	textSurface, err = t.Context.Font.RenderUTF8Blended(value,
		sdl.Color{
			R: t.Style.ForegroundColor.R,
			G: t.Style.ForegroundColor.G,
			B: t.Style.ForegroundColor.B,
			A: t.Style.ForegroundColor.A,
		},
	)
	if err != nil {
		panic(err)
	}
	defer textSurface.Free()
	textTexture, err = t.Context.Renderer.CreateTextureFromSurface(textSurface)
	if err != nil {
		panic(err)
	}
	defer textTexture.Destroy()

	if outlineTexture != nil {
		t.tw = outlineSurface.W
		t.th = outlineSurface.H
	} else {
		t.tw = textSurface.W
		t.th = textSurface.H
	}

	// Create our target texture.
	t.SDLTexture, err = t.Context.Renderer.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA32), sdl.TEXTUREACCESS_TARGET, t.tw, t.th)
	if err != nil {
		ShowError("%s", sdl.GetError())
		panic(err)
	}
	if err = t.SDLTexture.SetBlendMode(sdl.BLENDMODE_BLEND); err != nil {
		ShowError("%s", sdl.GetError())
		panic(err)
	}

	// Render our Text and Outline to it.
	oldTarget := t.Context.Renderer.GetRenderTarget()

	if err = t.Context.Renderer.SetRenderTarget(t.SDLTexture); err != nil {
		ShowError("%s", sdl.GetError())
		panic(err)
	}

	// Clear texture or we get errors (at least w/ intel gfx)
	t.Context.Renderer.SetDrawColor(0, 0, 0, 0)
	t.Context.Renderer.Clear()

	if outlineTexture != nil {
		if err = outlineTexture.SetAlphaMod(t.Style.OutlineColor.A); err != nil {
			fmt.Printf("%s\n", sdl.GetError())
			panic(err)
		}
		if err = t.Context.Renderer.Copy(outlineTexture, nil, nil); err != nil {
			panic(err)
		}
	}
	if textTexture != nil {
		dest := sdl.Rect{
			X: 0,
			Y: 0,
			W: textSurface.W,
			H: textSurface.H,
		}
		if outlineTexture != nil {
			dest.X = 2
			dest.Y = 2
		}
		if err = t.Context.Renderer.Copy(textTexture, nil, &dest); err != nil {
			panic(err)
		}
	}
	if err = t.Context.Renderer.SetRenderTarget(oldTarget); err != nil {
		panic(err)
	}

	t.Dirty = true
	t.OnChange()
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

// OnWindowResized is the same as BaseElement but always recreates the
// underlying SDL texture.
func (t *TextElement) OnWindowResized(w, h int32) {
	t.SetValue(t.Value)
	t.BaseElement.OnWindowResized(w, h)
}
