// +build !mobile

package ui

import (
	"image"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

// ImageElement is the element responsible for rendering an image.
type ImageElement struct {
	BaseElement
	SDLTexture *sdl.Texture
	Image      image.Image
	tw         int32 // Texture width
	th         int32 // Texture height
}

// Destroy destroys the underlying ImageElement.
func (i *ImageElement) Destroy() {
	if i.SDLTexture != nil {
		i.SDLTexture.Destroy()
	}
	i.BaseElement.Destroy()
}

// Render renders the ImageElement to the screen.
func (i *ImageElement) Render() {
	if i.IsHidden() {
		return
	}
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

// SetImage sets the underlying texture to the passed go Image.
func (i *ImageElement) SetImage(img image.Image) {
	if i.Context == nil {
		return
	}

	var err error
	var surface *sdl.Surface
	var bpp int
	var rmask, gmask, bmask, amask uint32
	var width, height int32

	width = int32(img.Bounds().Max.X)
	height = int32(img.Bounds().Max.Y)
	if bpp, rmask, gmask, bmask, amask, err = sdl.PixelFormatEnumToMasks(uint(sdl.PIXELFORMAT_RGBA32)); err != nil {
		panic(err)
	}
	// NOTE: It might be heavy to do these conversions each time SetImage is called. Perhaps
	// we should have SetImage only handle image.NRGBA and do any required load conversions
	// in data.Manager.
	switch img := img.(type) {
	case *image.NRGBA:
		surface, err = sdl.CreateRGBSurfaceFrom(
			unsafe.Pointer(&img.Pix[0]),
			width,
			height,
			bpp,
			img.Stride,
			rmask, gmask, bmask, amask)
	case *image.Paletted:
		bounds := img.Bounds()
		rgbaImage := image.NewNRGBA(bounds)
		for x := 0; x < bounds.Max.X; x++ {
			for y := 0; y < bounds.Max.Y; y++ {
				var pal = img.At(x, y)
				rgbaImage.Set(x, y, pal)
			}
		}
		surface, err = sdl.CreateRGBSurfaceFrom(
			unsafe.Pointer(&rgbaImage.Pix[0]),
			width,
			height,
			bpp,
			rgbaImage.Stride,
			rmask, gmask, bmask, amask)
	default:
		// FIXME: We really shouldn't just panic here.
		panic(err)
	}
	defer surface.Free()

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
