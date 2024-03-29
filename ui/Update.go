package ui

import "image/color"

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

// UpdateY represents a change to the Y of an Element.
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

// UpdateDimensions represents a change to the x, y, w, and h of an Element.
type UpdateDimensions struct {
	X, Y, W, H Number
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

// UpdateParseStyle parses the given style and incorporates it into the element.
type UpdateParseStyle string

// UpdateOutlineColor represents an update to the outline color of the element.
type UpdateOutlineColor = color.NRGBA

// UpdateBackgroundColor represents an update to the background color of the element.
type UpdateBackgroundColor color.NRGBA

// UpdateForegroundColor represents an update to the foreground color of the element.
type UpdateForegroundColor color.NRGBA

// UpdateImageID is used to do a lookup from the data manager to update an image element's image.
type UpdateImageID uint32

// UpdateHideImage is for hiding the rendering of an image element.
type UpdateHideImage = bool

// UpdateStyle represents an update to the Style of an element.
type UpdateStyle = string

// UpdateDirt is an message that marks the element as dirty or not.
type UpdateDirt = bool

// UpdateFocus is a message that marks the element as focused.
type UpdateFocus struct{}

// UpdateHidden as a message that marks the element to be hidden or not.
type UpdateHidden bool

// UpdateAlpha is a message that sets the element's alpha. Not applicable for all elements.
type UpdateAlpha = float64

// UpdateColorMod is a message that sets the element's color mod. Not applicable for all elements.
type UpdateColorMod color.NRGBA

// UpdateGrayscale is a message that sets the element's grayscale rendering. Not applicable for all elements.
type UpdateGrayscale bool
