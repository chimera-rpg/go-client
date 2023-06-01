//go:build !mobile
// +build !mobile

package ui

import (
	"image"

	"github.com/veandco/go-sdl2/sdl"
)

// ImageElement is the element responsible for rendering an image.
type ImageElement struct {
	BaseElement
	Textures       *Image
	OutlineTexture *sdl.Texture
	Image          image.Image
	ImageID        uint32
	hideImage      bool
	postOutline    bool
	grayscale      bool
	tw             int32 // Texture width
	th             int32 // Texture height
	invalidated    bool
}

// Destroy destroys the underlying ImageElement.
func (i *ImageElement) Destroy() {
	if i.ImageID == 0 {
		if i.Textures != nil {
			if i.Textures.regularTexture != nil {
				i.Textures.regularTexture.Destroy()
			}
			if i.Textures.grayscaleTexture != nil {
				i.Textures.grayscaleTexture.Destroy()
			}
		}
	}
	if i.OutlineTexture != nil {
		i.OutlineTexture.Destroy()
	}

	i.BaseElement.Destroy()
}

// Render renders the ImageElement to the screen.
func (i *ImageElement) Render() {
	if i.invalidated {
		if i.OutlineTexture != nil {
			i.OutlineTexture.Destroy()
			i.OutlineTexture = nil
		}
		i.UpdateOutline()
	}
	if i.IsHidden() || i.Image == nil {
		i.BaseElement.Render()
		return
	}
	if i.Textures == nil {
		i.SetImageID(i.ImageID)
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
			texture = i.Textures.grayscaleTexture
		} else {
			texture = i.Textures.regularTexture
		}
		if texture != nil {
			texture.SetColorMod(i.Style.ColorMod.R, i.Style.ColorMod.G, i.Style.ColorMod.B)
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

func (i *ImageElement) SetImageID(id uint32) {
	i.ImageID = id
	imgTextures := i.Context.Manager.GetImage(id)
	if imgTextures == nil {
		img, err := i.Context.Manager.GetCachedImage(id)
		i.Image = img
		/*if err != nil {
			panic(err)
		}*/
		tex, gray, err := i.Context.CreateTexture(img)
		if err != nil {
			panic(err)
		}
		i.Context.Manager.SetRegularTexture(id, tex)
		i.Context.Manager.SetGrayscaleTexture(id, gray)
		i.Context.Manager.imageTextures[id].width = int32(img.Bounds().Dx())
		i.Context.Manager.imageTextures[id].height = int32(img.Bounds().Dy())
		imgTextures = i.Context.Manager.GetImage(id)
	}
	if img, err := i.Context.Manager.GetCachedImage(id); err == nil {
		i.Image = img
	}

	i.Textures = imgTextures

	i.tw = imgTextures.width
	i.th = imgTextures.height
	if i.Style.Resize.Has(TOCONTENT) {
		w := float64(imgTextures.width)
		h := float64(imgTextures.height)
		if i.Style.ScaleX.Value > 0 {
			if i.Style.ScaleX.Percentage {
				w = i.Style.ScaleX.PercentOf(w)
			} else {
				w *= i.Style.ScaleX.Value
			}
		}
		if i.Style.ScaleY.Value > 0 {
			if i.Style.ScaleY.Percentage {
				h = i.Style.ScaleY.PercentOf(h)
			} else {
				h *= i.Style.ScaleY.Value
			}
		}

		i.Style.W.Set(w)
		i.Style.H.Set(h)
		i.CalculateStyle()
	}

	i.Dirty = true
	i.invalidated = true
}

func (i *ImageElement) UpdateOutline() {
	var err error
	if i.Style.OutlineColor.A > 0 {
		i.OutlineTexture, err = i.Context.CreateOutlineFromTexture(i.Textures.regularTexture, i.w, i.h, i.Style.OutlineColor)
		if err != nil {
			// TODO: Probably don't panic.
			panic(err)
		}
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
	if i.IsHidden() {
		return false
	}
	// Superior Image-based method since we have the image anyhow.
	x = x - i.ax
	y = y - i.ay
	rect := i.Image.Bounds()
	x1 := int(float64(x) * float64(rect.Dx()) / float64(int(i.w)))
	y1 := int(float64(y) * float64(rect.Dy()) / float64(int(i.h)))
	c := i.Image.At(x1, y1)
	_, _, _, a := c.RGBA()
	if a > 0 {
		return true
	}
	return false
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
	err = i.Context.Renderer.Copy(i.Textures.regularTexture, nil, nil)
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
	}
	return false*/
	// Resize-based method.
	/*x -= i.GetAbsoluteX()
	y -= i.GetAbsoluteY()

	resizedImage := resize.Resize(uint(i.w), uint(i.h), i.Image, resize.NearestNeighbor)
	_, _, _, a := resizedImage.At(int(x), int(y)).RGBA()

	if a > 0 {
		return true
	}

	return false*/
}
