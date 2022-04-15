package ui

// ElementI is the interface that all Element(s) are generally passed around
// and parented as.
type ElementI interface {
	// Handlers
	Destroy()
	Render()
	RenderPost()
	//
	GetAdoptChannel() chan ElementI
	GetDisownChannel() chan ElementI
	GetDestroyChannel() chan bool
	GetUpdateChannel() chan UpdateI
	HandleUpdate(UpdateI)
	//
	SetX(int32)
	GetX() int32
	SetY(int32)
	GetY() int32
	GetAbsoluteX() int32
	GetAbsoluteY() int32
	GetWidth() int32
	GetHeight() int32
	GetScrollLeft() int32
	GetScrollTop() int32
	GetZIndex() int
	//
	GetContext() *Context
	SetContext(c *Context)
	//
	// IsDirty returns if the Element should be redrawn
	SetDirty(bool)
	IsDirty() bool
	// HasDirt iterates down all of an element's children to see if any return true for IsDirty
	HasDirt() bool
	// Value is the Element's most obvious string field -- for Window it is the title, for Button it is the button text, for Text it is the contained text.
	SetValue(value string) error
	GetValue() string
	// Style is the Element's Styling related to color, size, and position.
	GetStyle() *Style
	// Calculates the given Element's style. Should be called whenever Style is changed.
	CalculateStyle()
	// ShouldRestyle indicates if the Element should have calculate style called on it.
	ShouldRestyle() bool
	//
	// Returns whether or not this container constrains x,y to be relative to itself
	IsContainer() bool
	SetParent(p ElementI)
	GetParent() ElementI
	AdoptChild(e ElementI)
	DisownChild(e ElementI)
	GetChildren() []ElementI
	//
	SetHidden(b bool)
	IsHidden() bool
	//
	Hit(x int32, y int32) bool
	PixelHit(x int32, y int32) bool
	// Events
	SetEvents(e Events)
	OnCreated()
	OnTouchBegin(id uint32, x int32, y int32) bool
	OnTouchMove(id uint32, x int32, y int32) bool
	OnTouchEnd(id uint32, x int32, y int32) bool
	OnMouseButtonDown(buttonID uint8, x int32, y int32) bool
	OnMouseMove(x int32, y int32) bool
	OnMouseIn(x int32, y int32) bool
	OnMouseOut(x int32, y int32) bool
	OnMouseButtonUp(buttonID uint8, x int32, y int32) bool
	OnPressed(buttonID uint8, x int32, y int32) bool
	OnHold(buttonID uint8, x int32, y int32) bool
	OnUnhold(buttonID uint8, x int32, y int32) bool
	OnKeyDown(key uint8, modifiers uint16, repeat bool) bool
	OnKeyUp(key uint8, modifiers uint16) bool
	OnTextInput(str string) bool
	OnTextEdit(str string, start int32, length int32) bool
	OnChange()
	OnAdopted(parent ElementI)
	CanFocus() bool
	SetFocused(bool)
	CanHold() bool
	SetHeld(bool)
	Focus()
	OnFocus() bool
	Blur()
	OnBlur() bool
	OnWindowResized(w, h int32)
	//
	IsGrayscale() bool
}

type BatchMessage interface{}

type BatchDestroyMessage struct {
	Target ElementI
}

type BatchAdoptMessage struct {
	Parent ElementI
	Target ElementI
}

type BatchDisownMessage struct {
	Parent ElementI
	Target ElementI
}

type BatchUpdateMessage struct {
	Target ElementI
	Update UpdateI
}
