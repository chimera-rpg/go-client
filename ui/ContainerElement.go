package ui

// ContainerRenderFunc is a type alias used for the custom render function callback.
type ContainerRenderFunc func(*Container)

// ContainerConfig is a configuration object that is used by NewContainer(c) or
// Container.Setup(c) to construct a functioning Container.
type ContainerConfig struct {
	Parent              *Container
	Style               string
	Events              Events
	ContainerRenderFunc ContainerRenderFunc
	Context             *Context
	Value               string
}

// ContainerElementStyle provides the default Style that is applied to all windows.
var ContainerElementStyle = `
	ForegroundColor 0 0 0 255
	BackgroundColor 139 186 139 255
	ScrollbarGripperColor 179 226 179 200
`

// NewContainerElement creates a new Container instance according to the passed ContainerConfig.
func NewContainerElement(c ContainerConfig) (w *Container, err error) {
	window := Container{}
	err = window.Setup(c)
	window.OnCreated()
	return &window, err
}

// Setup our window object according to the passed ContainerConfig.
func (w *Container) Setup(c ContainerConfig) (err error) {
	w.This = ElementI(w)
	w.SetupChannels()
	w.ContainerRenderFunc = c.ContainerRenderFunc
	w.Style.Parse(ContainerElementStyle)
	w.Style.Parse(c.Style)
	w.Events = c.Events
	w.Context = c.Context
	w.Value = c.Value
	w.SetDirty(true)

	// Trigger a resize so we can create a Texture
	//wid, err := w.SDLWindow.GetID()
	//w.Resize(wid, w.w, w.h)
	return nil
}

/*// GetX returns the cached x property. In the case of Containers this is 0.
func (w *Container) GetX() int32 {
	return 0
}

// GetY returns the cached y property. In the case of Containers this is 0.
func (w *Container) GetY() int32 {
	return 0
}*/

// IsContainer Returns whether or not this Element should be considered as a container.
func (w *Container) IsContainer() bool {
	return true
}

func (b *Container) refreshGrippers() {
	viewportH := float64(b.h)
	contentH := float64(b.h + b.overflowY)
	b.gripH = int32(viewportH / contentH * viewportH)
	if b.gripY+b.gripH > int32(viewportH) {
		b.gripY = int32(viewportH) - b.gripH
	} else if b.gripY < 0 {
		b.gripY = 0
	}
	b.Style.ScrollTop.Percentage = false
	b.Style.ScrollTop.Value = (float64(b.gripY) / viewportH) * contentH
	b.BaseElement.CalculateStyle()
	b.SetDirty(true)
}

func (b *Container) OnMouseMove(x int32, y int32) bool {
	if b.overflowY > 0 {
		if x >= (b.ax+b.w)-8 && x <= (b.ax+b.w) {
			b.gripHoverY = true
		} else {
			b.gripHoverY = false
		}
	}
	return b.BaseElement.OnMouseMove(x, y)
}

func (b *Container) OnMouseButtonDown(buttonID uint8, x int32, y int32) bool {
	if b.overflowY > 0 {
		if x >= (b.ax+b.w)-8 && x <= (b.ax+b.w) {
			ydiff := y - b.ay

			b.refreshGrippers()

			if ydiff >= b.gripY && ydiff <= b.gripY+b.gripH {
				b.gripHeldY = true
				b.gripLastY = ydiff
			} else {
				if ydiff < b.gripY {
					b.gripY -= b.gripH
				} else if ydiff > b.gripY+b.gripH {
					b.gripY += b.gripH
				}
				b.refreshGrippers()
			}
		}
	}
	return b.BaseElement.OnMouseButtonDown(buttonID, x, y)
}

func (b *Container) OnGlobalMouseMove(x int32, y int32) bool {
	if b.gripHeldY {
		ydiff := int32(y - b.ay)
		b.gripY += ydiff - b.gripLastY
		b.gripLastY = ydiff
		b.refreshGrippers()
	}
	return b.BaseElement.OnGlobalMouseMove(x, y)
}

func (b *Container) OnMouseWheel(x int32, y int32) bool {
	if b.overflowY > 0 {
		if y > 0 {
			b.gripY -= b.gripH
		} else if y < 0 {
			b.gripY += b.gripH
		}
		b.refreshGrippers()
	}
	return true
}

func (b *Container) OnGlobalMouseButtonUp(buttonID uint8, x int32, y int32) bool {
	if b.gripHeldY {
		b.gripHeldY = false
	}
	return b.BaseElement.OnGlobalMouseButtonUp(buttonID, x, y)
}
