// +build !MOBILE
package UI

import (
  "github.com/veandco/go-sdl2/sdl"
  "github.com/veandco/go-sdl2/ttf"
)

type Context struct {
  Renderer *sdl.Renderer
  Font *ttf.Font
}
