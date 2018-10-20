package UI

import (
  "github.com/veandco/go-sdl2/sdl"
)

type Button struct {
  BaseElement
  Text Text
  texture *sdl.Texture
}

func NewButton(value string, position sdl.Point) *Button {
  b := Button{}
  b.Color = sdl.Color{128, 128, 128, 255}
  b.Position = position
  b.SetValue(value)
  return &b
}

func (b *Button) Destroy() {
}

func (b *Button) Render(r *sdl.Renderer) {
  b.Text.Render(r)

  /*dst := sdl.Rect{
    X: 0,
    Y: 0,
    W: b.Size.X,
    H: b.Size.Y,
  }*/
}

func (b *Button) SetValue(value string) (err error) {
  err = b.Text.SetValue(value)
  return
}
