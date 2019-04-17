package UI

type ElementI interface {
	// Handlers
	Destroy()
	Render()
	//
	GetX() int32
	GetY() int32
	GetWidth() int32
	GetHeight() int32
	//
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
	//
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
	// Events
	OnTouchBegin(id uint32, x int32, y int32) bool
	OnTouchMove(id uint32, x int32, y int32) bool
	OnTouchEnd(id uint32, x int32, y int32) bool
	OnMouseButtonDown(button_id uint8, x int32, y int32) bool
	OnMouseMove(x int32, y int32) bool
	OnMouseButtonUp(button_id uint8, x int32, y int32) bool
}
