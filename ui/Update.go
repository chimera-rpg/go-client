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

// UpdateStyle represents an update to the Style of an element.
type UpdateStyle = string
