package UI

import (
  "github.com/veandco/go-sdl2/sdl"
  //"log"
)

type SpriteAtlas struct {
  texture *sdl.Texture
  Sprites map[int16]*Sprite
}

/*
//AddSprite adds the given byte slice to the texture atlas with the provided id mapping. It is presumed that the passed byte slice is actually a PNG file.
*/
func (s *SpriteAtlas) AddSprite(r *sdl.Renderer, mem []byte, id int16) (err error) {
  // TODO: Actually make this work as a texture atlas.
  /*rw, err := sdl.RWFromMem(mem)
  if err != nil {
    panic(err)
  }
  surface, err := img.LoadTypedRW(rw, true, "PNG")
  if err != nil {
    panic(err)
  }
  defer surface.Free()
  //
  if s.texture == nil {
    s.texture, err = r.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_STATIC, surface.W, surface.H)
    if err != nil {
      panic(err)
    }
  }
  //s.reformAtlas(r *sdl.Renderer, */
  s.Sprites[id], err = NewSpriteFromMem(r, mem)
  return
}
