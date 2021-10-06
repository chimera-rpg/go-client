package ui

// Events provide all the event handlers that may be implemented within an
// Element instance.
type Events struct {
	OnCreated         func()
	OnTouchBegin      func(id uint32, x int32, y int32) bool
	OnTouchMove       func(id uint32, x int32, y int32) bool
	OnTouchEnd        func(id uint32, x int32, y int32) bool
	OnMouseButtonDown func(button uint8, x int32, y int32) bool
	OnMouseMove       func(x int32, y int32) bool
	OnMouseButtonUp   func(button uint8, x int32, y int32) bool
	OnMouseIn         func(x int32, y int32) bool
	OnMouseOut        func(x int32, y int32) bool
	OnHold            func(button uint8, x int32, y int32) bool
	OnUnhold          func(button uint8, x int32, y int32) bool
	OnKeyDown         func(key uint8, modifiers uint16, repeat bool) bool
	OnKeyUp           func(key uint8, modifiers uint16) bool
	OnTextInput       func(str string) bool
	OnTextEdit        func(str string, start int32, length int32) bool
	OnTextSubmit      func(str string) bool
	OnChange          func() bool
	OnAdopted         func(parent ElementI)
	OnFocus           func() bool
	OnBlur            func() bool
	OnWindowResized   func(w int32, h int32)
}

type MouseEvent struct {
	Button uint8
	X, Y   int32
}
