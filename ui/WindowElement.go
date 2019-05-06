// +build !MOBILE

package ui

// RenderFunc is a type alias used for the custom render function callback.
type RenderFunc func(*Window)

// WindowConfig is a configuration object that is used by NewWindow(c) or
// Window.Setup(c) to construct a functioning Window.
type WindowConfig struct {
	Parent     *Window
	Style      string
	RenderFunc RenderFunc
	Context    *Context
	Value      string
}

// WindowElementStyle provides the default Style that is applied to all windows.
var WindowElementStyle = `
	ForegroundColor 0 0 0 255
	BackgroundColor 139 186 139 255
`

// NewWindow creates a new Window instance according to the passed WindowConfig.
func NewWindow(c WindowConfig) (w *Window, err error) {
	window := Window{}
	err = window.Setup(c)
	window.OnCreated()
	return &window, err
}

// GetX returns the cached x property. In the case of Windows this is 0.
func (w *Window) GetX() int32 {
	return 0
}

// GetY returns the cached y property. In the case of Windows this is 0.
func (w *Window) GetY() int32 {
	return 0
}

// IsContainer Returns whether or not this Element should be considered as a container.
func (w *Window) IsContainer() bool {
	return true
}
