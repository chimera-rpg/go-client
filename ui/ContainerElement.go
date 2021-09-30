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
