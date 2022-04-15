package ui

import "github.com/veandco/go-sdl2/sdl"

type Image struct {
	width, height    int32
	grayscaleTexture *sdl.Texture
	outlineTexture   *sdl.Texture
	regularTexture   *sdl.Texture
}
