package ui

// TextElementConfig is the configuration object passed to NewTextElement.
type TextElementConfig struct {
	Style  string
	Value  string
	Events Events
}

// TextElementStyle is our default styling for TextElements.
var TextElementStyle = `
	ForegroundColor 0 0 0 255
`

// NewTextElement creates a new TextElement from the passed configuration.
func NewTextElement(c TextElementConfig) ElementI {
	t := TextElement{}
	t.This = ElementI(&t)
	t.Style.Parse(TextElementStyle)
	t.Style.Parse(c.Style)
	t.SetValue(c.Value)
	t.Events = c.Events
	t.SetupChannels()

	t.OnCreated()

	return ElementI(&t)
}

// HandleUpdate is the base stub for handling update messages.
func (t *TextElement) HandleUpdate(update UpdateI) {
	switch u := update.(type) {
	case UpdateValue:
		t.SetValue(u.Value)
	default:
		t.BaseElement.HandleUpdate(update)
	}
}
