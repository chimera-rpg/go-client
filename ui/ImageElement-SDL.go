//go:build !mobile
// +build !mobile

package ui

import (
	"image"
	"math"
	"unsafe"

	"github.com/nfnt/resize"
	"github.com/veandco/go-sdl2/sdl"
)

// ImageElement is the element responsible for rendering an image.
type ImageElement struct {
	BaseElement
	SDLTexture       *sdl.Texture
	GrayscaleTexture *sdl.Texture
	OutlineTexture   *sdl.Texture
	Image            image.Image
	hideImage        bool
	postOutline      bool
	grayscale        bool
	tw               int32 // Texture width
	th               int32 // Texture height
}

// Destroy destroys the underlying ImageElement.
func (i *ImageElement) Destroy() {
	if i.SDLTexture != nil {
		i.SDLTexture.Destroy()
	}
	if i.OutlineTexture != nil {
		i.OutlineTexture.Destroy()
	}
	if i.GrayscaleTexture != nil {
		i.GrayscaleTexture.Destroy()
	}
	i.BaseElement.Destroy()
}

// Render renders the ImageElement to the screen.
func (i *ImageElement) Render() {
	if i.IsHidden() || i.Image == nil {
		i.BaseElement.Render()
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
	if !i.hideImage {
		var texture *sdl.Texture
		if i.grayscale {
			texture = i.GrayscaleTexture
		} else {
			texture = i.SDLTexture
		}
		if texture != nil {
			texture.SetAlphaMod(uint8(i.Style.Alpha.Value * 255))
			i.Context.Renderer.Copy(texture, nil, &dst)
			texture.SetAlphaMod(255)
		}
	}
	// Render outline.
	if !i.postOutline && i.OutlineTexture != nil {
		dst.X--
		dst.Y--
		dst.W += 2
		dst.H += 2
		i.Context.Renderer.Copy(i.OutlineTexture, nil, &dst)
	}
	i.BaseElement.Render()
}

// SetImage sets the underlying texture to the passed go Image.
func (i *ImageElement) SetImage(img image.Image) {
	if i.Context == nil || img == nil {
		return
	}
	i.Image = img

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
	// (re)create our outline if we should.
	if i.OutlineTexture != nil {
		i.OutlineTexture.Destroy()
		i.OutlineTexture = nil
	}
	i.UpdateOutline()
	if i.GrayscaleTexture != nil {
		i.GrayscaleTexture.Destroy()
		i.GrayscaleTexture = nil
		i.UpdateGrayscale()
	}

	i.Dirty = true
}

func (i *ImageElement) createGrayscale() error {
	texWidth := i.w
	texHeight := i.h

	tex, err := i.Context.Renderer.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA32), sdl.TEXTUREACCESS_TARGET, i.w, i.h)
	if err != nil {
		return err
	}
	defer tex.Destroy()
	prevRenderTarget := i.Context.Renderer.GetRenderTarget()
	defer i.Context.Renderer.SetRenderTarget(prevRenderTarget)
	i.Context.Renderer.SetRenderTarget(tex)
	i.Context.Renderer.SetDrawColor(0, 0, 0, 0)
	i.Context.Renderer.Clear()
	err = i.Context.Renderer.Copy(i.SDLTexture, nil, nil)
	if err != nil {
		return err
	}

	realPixels := make([]byte, texWidth*texHeight*4)
	err = i.Context.Renderer.ReadPixels(nil, uint32(sdl.PIXELFORMAT_RGBA32), unsafe.Pointer(&realPixels[0]), int(texWidth)*4)
	if err != nil {
		return err
	}

	// Create our final texture for outline rendering.
	realTex, err := i.Context.Renderer.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA32), sdl.TEXTUREACCESS_STATIC, texWidth, texHeight)
	realTex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		return err
	}
	targetPixels := make([]byte, texWidth*texHeight*4)

	for x := 0; x < int(texWidth); x++ {
		for y := 0; y < int(texHeight); y++ {
			t := (y*int(texWidth) + x) * 4
			r := realPixels[t]
			g := realPixels[t+1]
			b := realPixels[t+2]
			var v byte
			// Average
			// v = (r + g + b) / 3
			// Desaturation
			{
				v = byte((math.Max(float64(b), math.Max(float64(r), float64(g))) + math.Min(float64(b), math.Min(float64(r), float64(g)))) / 2)
			}
			// Minimum decomposition
			{
				//v = byte(math.Min(float64(b), math.Min(float64(r), float64(g))))
			}
			// Maximum decomposition
			/*{
				v = byte(math.Max(float64(b), math.Max(float64(r), float64(g))))
			}*/
			targetPixels[t] = v
			targetPixels[t+1] = v
			targetPixels[t+2] = v
			targetPixels[t+3] = realPixels[t+3]
		}
	}

	err = realTex.Update(nil, targetPixels, int(texWidth)*4)
	if err != nil {
		realTex.Destroy()
		return err
	}
	i.GrayscaleTexture = realTex

	return nil

}

func (i *ImageElement) createOutline() error {
	/*_, _, width, height, err := i.SDLTexture.Query()
	if err != nil {
		return err
	}
	// Add 2 pixels for guaranteed pixel borders.
	texWidth := width + 2
	texHeight := height + 2*/
	texWidth := i.w + 2
	texHeight := i.h + 2

	tex, err := i.Context.Renderer.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA32), sdl.TEXTUREACCESS_TARGET, i.w, i.h)
	if err != nil {
		return err
	}
	defer tex.Destroy()
	prevRenderTarget := i.Context.Renderer.GetRenderTarget()
	defer i.Context.Renderer.SetRenderTarget(prevRenderTarget)
	i.Context.Renderer.SetRenderTarget(tex)
	i.Context.Renderer.SetDrawColor(0, 0, 0, 0)
	i.Context.Renderer.Clear()
	err = i.Context.Renderer.Copy(i.SDLTexture, nil, &sdl.Rect{X: 1, Y: 1, W: i.w, H: i.h})
	if err != nil {
		return err
	}

	realPixels := make([]byte, texWidth*texHeight*4)
	err = i.Context.Renderer.ReadPixels(nil, uint32(sdl.PIXELFORMAT_RGBA32), unsafe.Pointer(&realPixels[0]), int(texWidth)*4)
	if err != nil {
		return err
	}

	// Create our final texture for outline rendering.
	realTex, err := i.Context.Renderer.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA32), sdl.TEXTUREACCESS_STATIC, texWidth, texHeight)
	realTex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		return err
	}
	targetPixels := make([]byte, texWidth*texHeight*4)

	for x := 0; x < int(texWidth); x++ {
		for y := 0; y < int(texHeight); y++ {
			t := (y*int(texWidth) + x) * 4
			if realPixels[t+3] == 0 { // Fully alpha
				hasNonAlphaNeighbor := false
				i2 := ((y+1)*int(texWidth) + x) * 4
				if i2 < len(realPixels) && realPixels[i2+3] > 0 {
					hasNonAlphaNeighbor = true
				}
				if !hasNonAlphaNeighbor {
					i2 = ((y-1)*int(texWidth) + x) * 4
					if i2 >= 0 && i2 < len(realPixels) && realPixels[i2+3] > 0 {
						hasNonAlphaNeighbor = true
					}
				}
				if !hasNonAlphaNeighbor {
					i2 = (y*int(texWidth) + x + 1) * 4
					if i2 < len(realPixels) && realPixels[i2+3] > 0 {
						hasNonAlphaNeighbor = true
					}
				}
				if !hasNonAlphaNeighbor {
					i2 = (y*int(texWidth) + x - 1) * 4
					if i2 >= 0 && i2 < len(realPixels) && realPixels[i2+3] > 0 {
						hasNonAlphaNeighbor = true
					}
				}
				if hasNonAlphaNeighbor {
					targetPixels[t] = i.Style.OutlineColor.R
					targetPixels[t+1] = i.Style.OutlineColor.G
					targetPixels[t+2] = i.Style.OutlineColor.B
					targetPixels[t+3] = i.Style.OutlineColor.A
				}
			}
		}
	}

	err = realTex.Update(nil, targetPixels, int(texWidth)*4)
	if err != nil {
		realTex.Destroy()
		return err
	}
	i.OutlineTexture = realTex

	return nil
}

func (i *ImageElement) UpdateGrayscale() {
	if i.grayscale && i.GrayscaleTexture == nil {
		i.createGrayscale()
	} else if i.GrayscaleTexture != nil {
		//i.GrayscaleTexture.Destroy()
		//i.GrayscaleTexture = nil
	}
}

func (i *ImageElement) UpdateOutline() {
	if i.Style.OutlineColor.A > 0 {
		i.createOutline()
	} else if i.OutlineTexture != nil {
		i.OutlineTexture.Destroy()
		i.OutlineTexture = nil
	}
}

func (i *ImageElement) RenderPost() {
	// Render outline.
	if i.postOutline && i.OutlineTexture != nil {
		dst := sdl.Rect{
			X: i.x + i.pl - 1,
			Y: i.y + i.pt - 1,
			W: i.w + 2,
			H: i.h + 2,
		}

		i.Context.Renderer.Copy(i.OutlineTexture, nil, &dst)
	}
	i.BaseElement.RenderPost()
}

func (i *ImageElement) PixelHit(x, y int32) bool {
	// Pure SDL texture method. Instable, probably have to lock pixels.
	/*texWidth := i.w
	texHeight := i.h

	tex, err := i.Context.Renderer.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA32), sdl.TEXTUREACCESS_TARGET, texWidth, texHeight)
	if err != nil {
		return false
	}
	defer tex.Destroy()

	prevRenderTarget := i.Context.Renderer.GetRenderTarget()
	defer i.Context.Renderer.SetRenderTarget(prevRenderTarget)

	i.Context.Renderer.SetRenderTarget(tex)
	i.Context.Renderer.SetDrawColor(0, 0, 0, 0)
	i.Context.Renderer.Clear()
	err = i.Context.Renderer.Copy(i.SDLTexture, nil, nil)
	if err != nil {
		return false
	}

	realPixels := make([]byte, texWidth*texHeight*4)
	err = i.Context.Renderer.ReadPixels(nil, uint32(sdl.PIXELFORMAT_RGBA32), unsafe.Pointer(&realPixels[0]), int(texWidth)*4)
	if err != nil {
		return false
	}

	x -= i.GetAbsoluteX()
	y -= i.GetAbsoluteY()

	t := (y*i.w + x) * 4
	if realPixels[t+3] > 0 {
		return true
	}*/
	// Resize-based method.
	x -= i.GetAbsoluteX()
	y -= i.GetAbsoluteY()

	resizedImage := resize.Resize(uint(i.w), uint(i.h), i.Image, resize.NearestNeighbor)
	_, _, _, a := resizedImage.At(int(x), int(y)).RGBA()

	if a > 0 {
		return true
	}

	return false
}
