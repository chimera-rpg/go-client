// +build !MOBILE

package ui

import (
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

// ImageElement is the element responsible for rendering an image.
type ImageElement struct {
	BaseElement
	SDLTexture *sdl.Texture
	Image      []byte
	tw         int32 // Texture width
	th         int32 // Texture height
}

// ImageElementConfig is the configuration for construction.
type ImageElementConfig struct {
	Image []byte
	Style string
}

// ImageElementStyle is our default style for ImageElement.
var ImageElementStyle = `
	ContentOrigin CenterX CenterY
`

// NewImageElement creates a new ImageElement from the passed configuration.
func NewImageElement(c ImageElementConfig) ElementI {
	i := ImageElement{}
	i.This = ElementI(&i)
	i.Style.Parse(ImageElementStyle)
	i.Style.Parse(c.Style)
	i.Image = c.Image

	i.OnCreated()

	return ElementI(&i)
}

// Destroy destroys the underlying ImageElement.
func (i *ImageElement) Destroy() {
	if i.SDLTexture != nil {
		i.SDLTexture.Destroy()
	}
}

// Render renders the ImageElement to the screen.
func (i *ImageElement) Render() {
	if i.IsHidden() {
		return
	}
	i.lock.Lock()
	defer i.lock.Unlock()
	if i.SDLTexture == nil {
		i.SetImage(i.Image)
	}
	if i.Style.BackgroundColor.A > 0 {
		dst := sdl.Rect{
			X: i.x,
			Y: i.y,
			W: i.w,
			H: i.h,
		}
		i.Context.Renderer.SetDrawColor(i.Style.BackgroundColor.R, i.Style.BackgroundColor.G, i.Style.BackgroundColor.B, i.Style.BackgroundColor.A)
		i.Context.Renderer.FillRect(&dst)
	}
	dst := sdl.Rect{
		X: i.x + i.pl,
		Y: i.y + i.pt,
		W: i.w,
		H: i.h,
	}
	i.Context.Renderer.Copy(i.SDLTexture, nil, &dst)
	i.BaseElement.Render()
}

// SetImage sets the underlying texture to the passed PNG byte slice.
func (i *ImageElement) SetImage(png []byte) {
	if i.Context == nil {
		return
	}

	rwops, err := sdl.RWFromMem(png)
	defer rwops.Close()
	surface, err := img.LoadRW(rwops, false)
	defer surface.Free()
	if err != nil {
		surface, err = sdl.CreateRGBSurface(0, 16, 16, 32, 0, 0, 0, 0)
		defer surface.Free()
		if err != nil {
			panic(err)
		}
	}
	i.SDLTexture, err = i.Context.Renderer.CreateTextureFromSurface(surface)
	if err != nil {
		panic(err)
	}
	i.tw = surface.W
	i.th = surface.H
	if i.Style.Resize.Has(TOCONTENT) {
		i.Style.W.Set(float64(surface.W))
		i.Style.H.Set(float64(surface.H))
	}
	i.Dirty = true
}
