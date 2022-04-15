//go:build !mobile
// +build !mobile

package ui

import (
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
