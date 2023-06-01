//go:build !mobile
// +build !mobile

package ui

import (
	"errors"
	"image"
	"image/color"
	"math"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// Context contains the current renderer, font, and other necessary information
// for rendering.
type Context struct {
	Renderer    *sdl.Renderer
	Manager     *DataManager
	Font        *ttf.Font
	OutlineFont *ttf.Font
}

// CreateTexture creates a regular and grayscale texture from the given image.
func (c *Context) CreateTexture(img image.Image) (tex *sdl.Texture, gray *sdl.Texture, err error) {
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

	tex, err = c.Renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return nil, nil, err
	}

	// Might as well also create the grayscale...
	{
		texWidth := int32(img.Bounds().Dx())
		texHeight := int32(img.Bounds().Dy())
		tempTex, err := c.Renderer.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA32), sdl.TEXTUREACCESS_TARGET, texWidth, texHeight)
		if err != nil {
			return nil, nil, err
		}
		defer tempTex.Destroy()
		prevRenderTarget := c.Renderer.GetRenderTarget()
		defer c.Renderer.SetRenderTarget(prevRenderTarget)
		c.Renderer.SetRenderTarget(tempTex)
		c.Renderer.SetDrawColor(0, 0, 0, 0)
		c.Renderer.Clear()
		err = c.Renderer.Copy(tex, nil, nil)
		if err != nil {
			return nil, nil, err
		}

		realPixels := make([]byte, texWidth*texHeight*4)
		err = c.Renderer.ReadPixels(nil, uint32(sdl.PIXELFORMAT_RGBA32), unsafe.Pointer(&realPixels[0]), int(texWidth)*4)
		if err != nil {
			return nil, nil, err
		}

		// Create our final texture for outline rendering.
		gray, err = c.Renderer.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA32), sdl.TEXTUREACCESS_STATIC, texWidth, texHeight)
		gray.SetBlendMode(sdl.BLENDMODE_BLEND)
		if err != nil {
			return nil, nil, err
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
		err = gray.Update(nil, targetPixels, int(texWidth)*4)
		if err != nil {
			gray.Destroy()
			return nil, nil, err
		}

	}

	return tex, gray, nil
}

func (c *Context) CreateOutlineFromTexture(texture *sdl.Texture, w, h int32, outline color.NRGBA) (*sdl.Texture, error) {
	if texture == nil {
		return nil, errors.New("missing textures to make outline")
	}
	/*_, _, width, height, err := i.SDLTexture.Query()
	if err != nil {
		return err
	}
	// Add 2 pixels for guaranteed pixel borders.
	texWidth := width + 2
	texHeight := height + 2*/

	texWidth := w + 2
	texHeight := h + 2

	tex, err := c.Renderer.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA32), sdl.TEXTUREACCESS_TARGET, w, h)
	if err != nil {
		return nil, err
	}
	defer tex.Destroy()
	prevRenderTarget := c.Renderer.GetRenderTarget()
	defer c.Renderer.SetRenderTarget(prevRenderTarget)
	c.Renderer.SetRenderTarget(tex)
	c.Renderer.SetDrawColor(0, 0, 0, 0)
	c.Renderer.Clear()
	err = c.Renderer.Copy(texture, nil, &sdl.Rect{X: 1, Y: 1, W: w, H: h})
	if err != nil {
		return nil, err
	}

	realPixels := make([]byte, texWidth*texHeight*4)
	err = c.Renderer.ReadPixels(nil, uint32(sdl.PIXELFORMAT_RGBA32), unsafe.Pointer(&realPixels[0]), int(texWidth)*4)
	if err != nil {
		return nil, err
	}

	// Create our final texture for outline rendering.
	realTex, err := c.Renderer.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA32), sdl.TEXTUREACCESS_STATIC, texWidth, texHeight)
	realTex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		return nil, err
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
					targetPixels[t] = outline.R
					targetPixels[t+1] = outline.G
					targetPixels[t+2] = outline.B
					targetPixels[t+3] = outline.A
				}
			}
		}
	}

	err = realTex.Update(nil, targetPixels, int(texWidth)*4)
	if err != nil {
		realTex.Destroy()
		return nil, err
	}
	return realTex, nil
}
