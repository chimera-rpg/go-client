package UI

import (
  "github.com/veandco/go-sdl2/sdl"
  "github.com/veandco/go-sdl2/ttf"
  //"log"
)

type Text struct {
  BaseElement
  Font *ttf.Font
  texture *sdl.Texture
  surface *sdl.Surface
}

func NewText(font *ttf.Font, value string, position sdl.Point) *Text {
  t := Text{
    BaseElement: BaseElement{
      Color: sdl.Color{255, 255, 255, 255},
    },
    Font: font,
  }
  t.Color = sdl.Color{255, 255, 255, 255}
  t.Position = position
  t.SetValue(value)
  return &t
}

func (t *Text) Destroy() {
  if t.texture != nil {
    t.texture.Destroy()
  }
}

func (t *Text) Render(r *sdl.Renderer) {
  var err error
  if t.Font == nil {
    return
  }
  if t.surface == nil {
    t.SetValue(t.Value)
  }
  if t.texture == nil {
    t.texture, err = r.CreateTextureFromSurface(t.surface)
  }
  if err != nil {
    panic(err)
  }
  dst := sdl.Rect{
    X: 0,
    Y: 0,
    W: t.Size.X,
    H: t.Size.Y,
  }
  r.Copy(t.texture, nil, &dst)
}

func (t *Text) SetValue(value string) (err error) {
  if t.Font == nil {
    return
  }
  t.Value = value
  if t.surface != nil {
    t.surface.Free()
  }
  if t.texture != nil {
    t.texture.Destroy()
    t.texture = nil
  }
  t.surface, err = t.Font.RenderUTF8Blended(t.Value, t.Color)
  t.Size.X = t.surface.W
  t.Size.Y = t.surface.H
  return
}
