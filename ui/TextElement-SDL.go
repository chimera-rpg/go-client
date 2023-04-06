//go:build !mobile
// +build !mobile

package ui

import (
	"fmt"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

// TextElement is our main element for handling and drawing text.
type TextElement struct {
	BaseElement
	SDLTexture *sdl.Texture
	tw         int32 // Texture width
	th         int32 // Texture height
	lines      []Line
}

type Line struct {
	x, y  int32
	w, h  int32
	value string
}

// Destroy handles the destruction of the underlying texture.
func (t *TextElement) Destroy() {
	if t.SDLTexture != nil {
		t.SDLTexture.Destroy()
	}
	t.BaseElement.Destroy()
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

	t.CalculateLines()

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

	oldTarget := t.Context.Renderer.GetRenderTarget()
	if err = t.Context.Renderer.SetRenderTarget(t.SDLTexture); err != nil {
		ShowError("%s", sdl.GetError())
		panic(err)
	}

	// Clear texture or we get errors (at least w/ intel gfx)
	t.Context.Renderer.SetDrawColor(0, 0, 0, 0)
	t.Context.Renderer.Clear()

	// Create text Outline
	for _, line := range t.lines {
		var textSurface, outlineSurface *sdl.Surface
		var textTexture, outlineTexture *sdl.Texture

		if t.Style.OutlineColor.A > 0 {
			outlineSurface, err = t.Context.OutlineFont.RenderUTF8Blended(line.value,
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
		textSurface, err = t.Context.Font.RenderUTF8Blended(line.value,
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
			dest.X += line.x
			dest.Y += line.y
			if err = t.Context.Renderer.Copy(textTexture, nil, &dest); err != nil {
				panic(err)
			}
		}
	}
	if err = t.Context.Renderer.SetRenderTarget(oldTarget); err != nil {
		panic(err)
	}

	t.Dirty = true
	t.OnChange()
	return
}

// GetFittedLines returns the text value as a series of strings that fit within the element's parent.
func (t *TextElement) GetFittedLines() []Line {
	// This is really bad.
	containerW := t.Parent.GetWidth()
	closestCellW, closestCellH, err := t.Context.Font.SizeUTF8("A")
	if err != nil {
		panic(err)
	}
	vagueLineWidth := containerW / int32(closestCellW)
	y := int32(0)
	var lines []Line

	splitLines := strings.Split(t.Value, "\n")
	for _, l := range splitLines {
		lastPos := int32(0)
		if len(l) == 0 {
			lines = append(lines, Line{
				x:     0,
				y:     y,
				w:     int32(closestCellW),
				h:     int32(closestCellW),
				value: " ",
			})
			y += int32(closestCellH)
		} else {
			for i := 0; i < len(l); i++ {
				currentPos := lastPos + vagueLineWidth
				if currentPos >= int32(len(l)) {
					currentPos = int32(len(l))
				} else {
					for l[currentPos] != ' ' && currentPos > 0 {
						currentPos--
					}
				}

				potentialLine := l[lastPos:currentPos]
				potentialLineWidth, potentialLineHeight, _ := t.Context.Font.SizeUTF8(potentialLine)

				for potentialLineWidth > int(containerW) {
					currentPos--

					for l[currentPos] != ' ' && currentPos > 0 {
						currentPos--
					}

					potentialLine = l[lastPos:currentPos]
					potentialLineWidth, potentialLineHeight, _ = t.Context.Font.SizeUTF8(potentialLine)
				}
				potentialLine = strings.TrimPrefix(potentialLine, " ")

				if len(potentialLine) == 0 {
					potentialLine = " "
					potentialLineHeight = closestCellH
				}
				lastPos = currentPos
				lines = append(lines, Line{
					x:     0,
					y:     y,
					w:     int32(potentialLineWidth),
					h:     int32(potentialLineHeight),
					value: potentialLine,
				})
				y += int32(potentialLineHeight)
				i = int(currentPos)
			}
		}
	}
	return lines
}

func (t *TextElement) CalculateLines() {
	if t.Style.Wrap.Has(WRAP) {
		t.lines = t.GetFittedLines()
	} else {
		w, h, err := t.Context.Font.SizeUTF8(t.Value)
		if err != nil {
			panic(err)
		}
		value := t.Value
		if value == "" {
			value = " "
		}
		t.lines = []Line{{
			value: value,
			x:     0,
			y:     0,
			w:     int32(w),
			h:     int32(h),
		}}
	}

	// FIXME: Store this during GetFittedLines
	var w int32
	for _, l := range t.lines {
		if l.w > w {
			w = l.w
		}
	}

	h := t.lines[len(t.lines)-1].y + t.lines[len(t.lines)-1].h

	if t.Style.OutlineColor.A > 0 {
		w += 4
		h += 4
	}

	if w == 0 {
		w = 1
	}
	if h == 0 {
		h = 1
	}

	t.tw = w
	t.th = h
	t.w = int32(w)
	t.h = int32(h)
	// FIXME: We shouldn't do this.
	t.Style.W.Percentage = false
	t.Style.W.Value = float64(w)
	t.Style.H.Percentage = false
	t.Style.H.Value = float64(h)
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
