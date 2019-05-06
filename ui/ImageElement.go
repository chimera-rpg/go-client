package ui

// ImageElementConfig is the configuration for construction.
type ImageElementConfig struct {
	Image []byte
	Style string
}

// ImageElementStyle is our default style for ImageElement.
var ImageElementStyle = `
	ContentOrigin CenterX CenterY
`

// NewImageElement creates a new ImageElement from the passed configuration.
func NewImageElement(c ImageElementConfig) ElementI {
	i := ImageElement{}
	i.This = ElementI(&i)
	i.Style.Parse(ImageElementStyle)
	i.Style.Parse(c.Style)
	i.Image = c.Image
	i.SetupChannels()

	i.OnCreated()

	return ElementI(&i)
}