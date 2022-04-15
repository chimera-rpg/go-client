package ui

import (
	"image"
)

// ImageElementConfig is the configuration for construction.
type ImageElementConfig struct {
	Image       image.Image
	ImageID     uint32
	Style       string
	Events      Events
	HideImage   bool
	PostOutline bool
	Grayscale   bool
}

// ImageElementStyle is our default style for ImageElement.
var ImageElementStyle = `
	ContentOrigin CenterX CenterY
`

// NewImageElement creates a new ImageElement from the passed configuration.
func NewImageElement(c ImageElementConfig) ElementI {
	i := ImageElement{}
	i.This = ElementI(&i)
	i.Style.Alpha.Set(1)
	i.Style.ColorMod.R = 255
	i.Style.ColorMod.G = 255
	i.Style.ColorMod.B = 255
	i.Style.Parse(ImageElementStyle)
	i.Style.Parse(c.Style)
	if c.Image != nil {
		i.Image = c.Image
	}
	if c.ImageID > 0 {
		i.ImageID = c.ImageID
	}
	i.hideImage = c.HideImage
	i.postOutline = c.PostOutline
	i.grayscale = c.Grayscale
	i.Events = c.Events
	i.SetupChannels()

	i.OnCreated()

	return ElementI(&i)
}

// HandleUpdate is the method for handling update messages.
func (i *ImageElement) HandleUpdate(update UpdateI) {
	switch u := update.(type) {
	case UpdateImageID:
		img, _ := i.Context.Manager.GetCachedImage(uint32(u))
		i.SetImage(img)
		i.OnChange()
	case image.Image:
		i.SetImage(u)
		i.OnChange()
	case UpdateOutlineColor:
		i.BaseElement.HandleUpdate(update)
		i.UpdateOutline()
		i.OnChange()
	case UpdateHideImage:
		i.hideImage = u
	case UpdateGrayscale:
		i.grayscale = bool(u)
		i.UpdateGrayscale()
	default:
		i.BaseElement.HandleUpdate(update)
	}
}

func (i *ImageElement) IsGrayscale() bool {
	return i.grayscale
}
