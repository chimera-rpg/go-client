// +build !MOBILE
package UI

import (
  "github.com/veandco/go-sdl2/sdl"
  "github.com/veandco/go-sdl2/img"
)

type ImageStyle struct {
  Style
  Scale float32
}

type ImageElement struct {
  BaseElement
  SDL_texture *sdl.Texture
  Image []byte
  tw int32 // Texture width
  th int32 // Texture height
  Scale float32
}

type ImageElementConfig struct {
  Image []byte
  Style ImageStyle
}

func NewImageElement(c ImageElementConfig) ElementI {
  i := ImageElement{}
  i.This  = ElementI(&i)
  i.Style = c.Style.Style
  i.Scale = c.Style.Scale
  i.Image = c.Image

  return ElementI(&i)
}

func (i *ImageElement) Destroy() {
}

func (i *ImageElement) Render() {
  if i.IsHidden() {
    return
  }
  if i.SDL_texture == nil {
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
    W: int32(float32(i.tw)*i.Scale),
    H: int32(float32(i.th)*i.Scale),
  }
  i.Context.Renderer.Copy(i.SDL_texture, nil, &dst)
  i.BaseElement.Render()
}

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
  i.SDL_texture, err = i.Context.Renderer.CreateTextureFromSurface(surface)
  if err != nil {
    panic(err)
  }
  i.tw = surface.W
  i.th = surface.H
  i.Style.W.Value = float64(surface.W)
  i.Style.H.Value = float64(surface.H)
  i.Dirty = true
}
