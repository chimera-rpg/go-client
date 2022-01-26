//go:build !mobile
// +build !mobile

package ui

import (
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
)

// PrimitiveElement is the element responsible for rendering a primitive.
type PrimitiveElement struct {
	BaseElement
	Shape PrimitiveShape
}

func (p *PrimitiveElement) Render() {
	if p.IsHidden() {
		return
	}
	if p.Shape == RectangleShape {
		// Draw filled box
		dst := sdl.Rect{
			X: p.x,
			Y: p.y,
			W: p.w,
			H: p.h,
		}
		if p.Style.BackgroundColor.A > 0 {
			p.Context.Renderer.SetDrawColor(
				p.Style.BackgroundColor.R,
				p.Style.BackgroundColor.G,
				p.Style.BackgroundColor.B,
				p.Style.BackgroundColor.A,
			)
			p.Context.Renderer.FillRect(&dst)
		}
		// Draw outline
		if p.Style.OutlineColor.A > 0 {
			p.Context.Renderer.SetDrawColor(
				p.Style.OutlineColor.R,
				p.Style.OutlineColor.G,
				p.Style.OutlineColor.B,
				p.Style.OutlineColor.A,
			)
			p.Context.Renderer.DrawRect(&dst)

		}
	} else if p.Shape == EllipseShape {
		// Draw filled shape.
		if p.Style.BackgroundColor.A > 0 {
			gfx.FilledEllipseRGBA(
				p.Context.Renderer,
				p.x+p.w/2,
				p.y+p.h/2,
				p.w/2,
				p.h/2,
				p.Style.BackgroundColor.R,
				p.Style.BackgroundColor.G,
				p.Style.BackgroundColor.B,
				p.Style.BackgroundColor.A,
			)
		}
		// Draw outline.
		if p.Style.OutlineColor.A > 0 {
			gfx.EllipseRGBA(
				p.Context.Renderer,
				p.x+p.w/2,
				p.y+p.h/2,
				p.w/2,
				p.h/2,
				p.Style.OutlineColor.R,
				p.Style.OutlineColor.G,
				p.Style.OutlineColor.B,
				p.Style.OutlineColor.A,
			)
		}
	}
	p.BaseElement.Render()
}
