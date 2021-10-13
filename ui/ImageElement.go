package ui

import (
	"image"
)

// ImageElementConfig is the configuration for construction.
type ImageElementConfig struct {
	Image     image.Image
	Style     string
	Events    Events
	HideImage bool
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
	if c.Image != nil {
		i.Image = c.Image
	}
	i.hideImage = c.HideImage
	i.Events = c.Events
	i.SetupChannels()

	i.OnCreated()

	return ElementI(&i)
}

// HandleUpdate is the method for handling update messages.
func (i *ImageElement) HandleUpdate(update UpdateI) {
	switch u := update.(type) {
	case image.Image:
		i.SetImage(u)
		i.OnChange()
	case UpdateOutlineColor:
		i.BaseElement.HandleUpdate(update)
		i.UpdateOutline()
		i.OnChange()
	case UpdateHideImage:
		i.hideImage = u
	default:
		i.BaseElement.HandleUpdate(update)
	}
}
