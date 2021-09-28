package ui

// UpdateI is the interface for our element Update messages.
type UpdateI interface {
}

// UpdateValue represents a change to the Value of an Element.
type UpdateValue struct {
	Value string
}

// UpdateX represents a change to the X of an Element.
type UpdateX struct {
	Number
}

// UpdateY represents a change to the X of an Element.
type UpdateY struct {
	Number
}

// UpdateW represents a change to the W of an Element.
type UpdateW struct {
	Number
}

// UpdateH represents a change to the H of an Element.
type UpdateH struct {
	Number
}

// UpdateScroll represents a change to the scroll left and top of an Element.
type UpdateScroll struct {
	Left, Top Number
}

// UpdateScrollLeft represents a change to the scroll left of an Element.
type UpdateScrollLeft struct {
	Number
}

// UpdateScrollTop represents a change to the scroll top of an Element.
type UpdateScrollTop struct {
	Number
}

// UpdateZIndex represents a change to the ZIndex of an Element.
type UpdateZIndex struct {
	Number
}

// UpdateStyle represents an update to the Style of an element.
type UpdateStyle = string

// UpdateDirt is an message that marks the element as dirty or not.
type UpdateDirt = bool
