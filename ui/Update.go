package ui

// UpdateI is the interface for our element Update messages.
type UpdateI interface {
}

// UpdateValue represents a change to the Value of an Element.
type UpdateValue struct {
	Value string
}
