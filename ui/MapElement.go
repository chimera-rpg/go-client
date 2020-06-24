package ui

// MapElementConfig is the configuration for construction.
type MapElementConfig struct {
	Style string
}

// MapElementStyle is our default style for MapElement.
var MapElementStyle = `
	ContentOrigin CenterX CenterY
`

// NewMapElement creates a new MapElement from the passed configuration.
func NewMapElement(c MapElementConfig) ElementI {
	i := MapElement{}
	i.This = ElementI(&i)
	i.Style.Parse(MapElementStyle)
	i.Style.Parse(c.Style)
	i.SetupChannels()

	i.OnCreated()

	return ElementI(&i)
}

// UpdateMap represents a change to the Value of an Element.
type UpdateMap struct {
	Height, Width, Depth int
}
