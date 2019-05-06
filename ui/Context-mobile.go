// +build mobile

package ui

import (
	"golang.org/x/mobile/gl"
)

// Context contains the current renderer, font, and other necessary information
// for rendering.
type Context struct {
	GLContext     gl.Context
	Width, Height int
	PixelRatio    float32
}
