package UI

import (
  "github.com/veandco/go-sdl2/sdl"
  "sync"
)

type RenderFunc func(*Window)

type WindowConfig struct {
  Window *Window
  Dimensions sdl.Rect
  ParentDimensions sdl.Rect
  RenderFunc RenderFunc
  Title string
}

type Window struct {
  Window *sdl.Window
  Renderer *sdl.Renderer
  Texture *sdl.Texture
  Dimensions sdl.Rect
  Title string
  RenderFunc RenderFunc
  ownsWindow bool
  Parent *Window
  ParentDimensions sdl.Rect
  Children []*Window
  Elements []*Element
  RenderMutex sync.Mutex
}

func NewWindow(c WindowConfig) (w *Window, err error) {
  window := Window{}
  err = window.Setup(c)
  return &window, err
}

func (w *Window) Setup(c WindowConfig) (err error) {
  w.RenderMutex = sync.Mutex{}
  // If window is nil, we create our own.
  w.Title = c.Title
  w.RenderFunc = c.RenderFunc
  if c.Window != nil {
    w.Window = c.Window.Window
    w.ownsWindow = false
    c.Window.AddChild(w)
  } else {
    w.Window, err = sdl.CreateWindow(c.Title, c.Dimensions.X, c.Dimensions.Y, c.Dimensions.W, c.Dimensions.H, sdl.WINDOW_SHOWN | sdl.WINDOW_RESIZABLE)
    w.Title = c.Title
    w.ownsWindow = true
  }
  if err != nil {
    return err
  }
  w.ParentDimensions = c.ParentDimensions
  w.Dimensions = c.Dimensions
  // Create our Renderer
  w.Renderer, err = w.Window.GetRenderer()
  if w.Renderer == nil {
    w.Renderer, err = sdl.CreateRenderer(w.Window, -1, sdl.RENDERER_ACCELERATED)
  }
  if err != nil {
    return err
  }
  // Trigger a resize so we can create a Texture
  wid, err := w.Window.GetID()
  w.Resize(wid, c.Dimensions.W, c.Dimensions.H)
  if err != nil {
    return err
  }
  return nil
}
func (w *Window) Destroy() {
  w.RemoveFromParent()
  if (w.ownsWindow) {
    w.Window.Destroy()
    w.Renderer.Destroy()
  }
  if w.Texture != nil {
    w.Texture.Destroy()
  }
  for _, child := range w.Children {
    child.Destroy()
  }
}
func (w *Window) Resize(id uint32, width int32, height int32) (err error) {
  wid, err := w.Window.GetID()
  if wid == id {
    if w.Parent != nil {
      if w.ParentDimensions.W != 0 {
        width = int32(float32(w.ParentDimensions.W) * 0.01 * float32(w.Parent.Dimensions.W))
        w.Dimensions.X = int32(float32(w.ParentDimensions.X) * 0.01 * float32(w.Parent.Dimensions.W))
      }
      if w.ParentDimensions.H != 0 {
        height = int32(float32(w.ParentDimensions.H) * 0.01 * float32(w.Parent.Dimensions.H))
        w.Dimensions.Y = int32(float32(w.ParentDimensions.Y) * 0.01 * float32(w.Parent.Dimensions.H))
      }
    }
    if w.ownsWindow != true {
      newTarget, err := w.Renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, width, height)
      if err != nil {
        return err
      }
      if w.Texture != nil {
        w.Renderer.SetRenderTarget(newTarget)
        dst := sdl.Rect{0, 0, w.Dimensions.W, w.Dimensions.H}
        w.Renderer.Copy(w.Texture, nil, &dst)
        w.Texture.Destroy()
      }
      w.Renderer.SetRenderTarget(newTarget)
      w.Texture = newTarget
    } else {
      w.Window.SetSize(width, height)
    }
    w.Dimensions.W = width
    w.Dimensions.H = height
  }
  for _, child := range w.Children {
    child.Resize(id, width, height)
  }
  return nil
}
func (w *Window) Render() {
  w.Renderer.SetRenderTarget(w.Texture)
  if w.RenderFunc != nil {
    w.RenderFunc(w)
  }
  for _, element := range w.Elements {
    (*element).Render(w.Renderer)
  }
  for _, child := range w.Children {
    child.Render()
    w.Renderer.SetRenderTarget(w.Texture)
    w.Renderer.Copy(child.Texture, nil, &child.Dimensions)
  }
  if w.ownsWindow == true {
    w.Renderer.Present()
  }
}

func (w *Window) AddChild(child *Window) {
  w.Children = append(w.Children, child)
  child.Parent = w
}
func (w *Window) RemoveFromParent() {
  if w.Parent != nil {
    for i, child := range w.Parent.Children {
      if child == w {
        w.Parent.Children = append(w.Parent.Children[:i], w.Parent.Children[i+1:]...)
        w.Parent = nil
        return
      }
    }
  }
}

func (w *Window) AddElement(el *Element) {
  w.Elements = append(w.Elements, el)
}
func (w *Window) RemoveElement(target *Element) {
  for i, el := range w.Elements {
    if el == target {
      w.Elements = append(w.Elements[:i], w.Elements[i+1:]...)
      return
    }
  }
}
