package ui

// PrimitiveShape represents our enum for shapes.
type PrimitiveShape int

//
const (
	RectangleShape PrimitiveShape = iota
	EllipseShape
)

// PrimitiveElementConfgi is the configuration for construction.
type PrimitiveElementConfig struct {
	Style  string
	Shape  PrimitiveShape
	Events Events
}

// PrimitiveElementStyle is our default style for PrimitiveElement.
var PrimitiveElementStyle = `
`

// NewPrimitiveElement creates a new PrimitiveEllemnt from the passed configuration.
func NewPrimitiveElement(c PrimitiveElementConfig) ElementI {
	p := PrimitiveElement{}
	p.This = ElementI(&p)
	p.Style.Alpha.Set(1)
	p.Style.Parse(PrimitiveElementStyle)
	p.Style.Parse(c.Style)
	p.Shape = c.Shape
	p.Events = c.Events
	p.SetupChannels()
	p.OnCreated()

	return ElementI(&p)
}
