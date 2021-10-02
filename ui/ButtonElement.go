package ui

// ButtonElementConfig provides the configuration for a new ButtonElemenb.
type ButtonElementConfig struct {
	Style   string
	Value   string
	Events  Events
	NoFocus bool
	NoHold  bool
}

// ButtonElementStyle is the default style for ButtonElements.
var ButtonElementStyle = `
	ForegroundColor 255 255 255 255
	BackgroundColor 139 139 186 128
	ContentOrigin CenterX CenterY
	MinH 12
	H 7%
	MaxH 40
`

// NewButtonElement creates a ButtonElement using the passed configuration.
func NewButtonElement(c ButtonElementConfig) ElementI {
	b := ButtonElement{}
	b.This = ElementI(&b)
	if c.NoHold {
		b.Holdable = false
	} else {
		b.Holdable = true
	}
	if c.NoFocus {
		b.Focusable = false
	} else {
		b.Focusable = true
	}
	b.Style.Parse(ButtonElementStyle)
	b.Style.Parse(c.Style)
	b.SetValue(c.Value)
	b.Events = c.Events
	b.SetupChannels()
	b.OnCreated()

	return ElementI(&b)
}

// OnKeyDown sets the button's held state when the enter key is pressed.
func (b *ButtonElement) OnKeyDown(key uint8, modifiers uint16, repeat bool) bool {
	switch key {
	case 13: // Activate button when enter is hit
		if b.CanHold() {
			b.SetHeld(true)
		}
	}
	return false
}

// OnKeyUp unsets the button's held state and triggers the OnMouseButtonUp
// method when the enter key is released.
func (b *ButtonElement) OnKeyUp(key uint8, modifiers uint16) bool {
	switch key {
	case 13: // Activate button when enter is released
		if b.CanHold() {
			b.SetHeld(false)
			b.OnMouseButtonUp(1, 0, 0)
		}
	}
	return false
}

// HandleUpdate is the method for handling update messages.
func (b *ButtonElement) HandleUpdate(update UpdateI) {
	switch u := update.(type) {
	case UpdateValue:
		b.SetValue(u.Value)
	}
}
