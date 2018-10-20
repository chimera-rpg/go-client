package UI

import (
  "github.com/veandco/go-sdl2/sdl"
)

type Element interface {
  //Press()
  Destroy()
  SetColors(color sdl.Color)
  GetColors() sdl.Color
  SetDimensions(rect sdl.Rect)
  GetDimensions() sdl.Rect
  SetValue(value string) error
  GetValue() string
  HandlePress(button uint8, state bool, x int, y int) bool
  Render(r *sdl.Renderer)
}

type BaseElement struct {
  Parent *Element
  Children []*Element
  Color sdl.Color
  Position sdl.Point
  Size sdl.Point
  Value string
}

func (b *BaseElement) HandlePress(button uint8, state bool, x int, y int) bool {
  return false
}

func (b *BaseElement) SetColors(color sdl.Color) {
  b.Color = color
}
func (b *BaseElement) GetColors() sdl.Color {
  return b.Color
}

func (b *BaseElement) GetDimensions() sdl.Rect {
  return sdl.Rect{
    X: b.Position.X,
    Y: b.Position.Y,
    W: b.Size.X,
    H: b.Size.Y,
  }
}
func (b *BaseElement) SetDimensions(rect sdl.Rect) {
  b.Position = sdl.Point{
    X: rect.X,
    Y: rect.Y,
  }
  b.Size = sdl.Point{
    X: rect.W,
    Y: rect.H,
  }
}

func (b *BaseElement) SetValue(value string) (err error) {
  b.Value = value
  return
}
func (b *BaseElement) GetValue() string {
  return b.Value
}
