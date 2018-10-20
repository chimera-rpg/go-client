package UI

import (
  "github.com/veandco/go-sdl2/sdl"
  "github.com/veandco/go-sdl2/img"
  //"log"
)

type Sprite struct {
  BaseElement
  texture *sdl.Texture
}

func NewSpriteFromMem(r *sdl.Renderer, mem []byte) (s *Sprite, err error) {
  rw, err := sdl.RWFromMem(mem)
  if err != nil {
    panic(err)
  }
  surface, err := img.LoadTypedRW(rw, true, "PNG")
  if err != nil {
    panic(err)
  }
  defer surface.Free()
  texture, err := r.CreateTextureFromSurface(surface)
  if err != nil {
    panic(err)
  }
  s = NewSprite(texture, &sdl.Rect{0, 0, surface.W, surface.H})
  return
}

func NewSprite(texture *sdl.Texture, clip *sdl.Rect) *Sprite {
  s := Sprite{}
  s.Position = sdl.Point{
    X: clip.X,
    Y: clip.Y,
  }
  s.Size = sdl.Point{
    X: clip.W,
    Y: clip.H,
  }
  return &s
}

func (s *Sprite) Destroy() {
}

func (s *Sprite) RenderAt(r *sdl.Renderer, x int32, y int32) (err error) {
  src := sdl.Rect{
    X: s.Position.X,
    Y: s.Position.Y,
    W: s.Size.X,
    H: s.Size.Y,
  }
  dst := sdl.Rect{
    X: x,
    Y: y,
    W: s.Size.X,
    H: s.Size.Y,
  }
  r.Copy(s.texture, &src, &dst)
  return
}
