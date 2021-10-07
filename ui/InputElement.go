package ui

// InputElementConfig is the construction configuration for an InputElement.
type InputElementConfig struct {
	Style         string
	Value         string
	Events        Events
	Password      bool
	Placeholder   string
	SubmitOnEnter bool
	ClearOnSubmit bool
	BlurOnSubmit  bool
}

// InputElementStyle is the default styling for an InputElement.
var InputElementStyle = `
	ForegroundColor 255 255 255 255
	BackgroundColor 0 0 0 128
	Padding 6
	ContentOrigin CenterY
	MinH 12
	H 7%
	MaxH 30
`

// NewInputElement creates a new InputElement using the passed configuration.
func NewInputElement(c InputElementConfig) ElementI {
	i := InputElement{}
	i.This = ElementI(&i)
	i.Style.Parse(InputElementStyle)
	i.Style.Parse(c.Style)
	i.composition = []rune(c.Value)
	i.cursor = len(i.composition)
	i.SyncComposition()
	i.Events = c.Events
	i.isPassword = c.Password
	i.placeholder = c.Placeholder
	i.Focusable = true
	i.submitOnEnter = c.SubmitOnEnter
	i.clearOnSubmit = c.ClearOnSubmit
	i.blurOnSubmit = c.BlurOnSubmit
	i.keysHeld = make(map[uint8]bool)
	i.SetupChannels()

	i.OnCreated()

	return ElementI(&i)
}

// SyncComposition is used to synchronize the element's value with the
// current composition.
func (i *InputElement) SyncComposition() {
	i.SetValue(string(i.composition))
}

// ClearComposition clears the current composition.
func (i *InputElement) ClearComposition() {
	i.SetValue("")
	i.composition = []rune("")
	i.cursor = 0
}

// OnKeyDown handles base key presses for moving the cursor, deleting runes, and
// otherwise.
func (i *InputElement) OnKeyDown(key uint8, modifiers uint16, repeat bool) bool {
	if !i.Focused {
		return true
	}
	i.keysHeld[key] = true
	switch key {
	case 27: // esc
		//BlurFocusedElement()
	case 8: // backspace
		if i.cursor > 0 {
			i.composition = append(i.composition[:i.cursor-1], i.composition[i.cursor:]...)
			i.cursor--
		}
	case 127: // delete
		if i.cursor < len(i.composition) {
			i.composition = append(i.composition[:i.cursor], i.composition[i.cursor+1:]...)
		}
	case 9: // tab
	case 79: // right
		i.cursor++
		if i.cursor > len(i.composition) {
			i.cursor = len(i.composition)
		}
	case 80: // left
		i.cursor--
		if i.cursor < 0 {
			i.cursor = 0
		}
	case 81: // down
		i.cursor = 0
	case 82: // up
		i.cursor = len(i.composition)
	}
	i.SyncComposition()
	if i.Events.OnKeyDown != nil {
		return i.Events.OnKeyDown(key, modifiers, repeat)
	}
	return true
}

// OnKeyUp handles base key releases.
func (i *InputElement) OnKeyUp(key uint8, modifiers uint16) bool {
	switch key {
	case 13: // enter
		if i.keysHeld[key] {
			if i.submitOnEnter {
				i.OnTextSubmit(string(i.composition))
			}
			if i.clearOnSubmit {
				i.ClearComposition()
			}
			if i.blurOnSubmit {
				i.Blur()
			}
		}
	}
	i.keysHeld[key] = false
	if i.Events.OnKeyUp != nil {
		return i.Events.OnKeyUp(key, modifiers)
	}
	return true
}

// OnTextInput handles the input of complete runes and appends them to the
// composition according to the cursor positining.
func (i *InputElement) OnTextInput(str string) bool {
	runes := []rune(str)
	i.composition = append(i.composition[:i.cursor], append(runes, i.composition[i.cursor:]...)...)
	i.cursor += len(runes)
	i.SyncComposition()
	if i.Events.OnTextInput != nil {
		return i.Events.OnTextInput(str)
	}
	return true
}

// OnTextEdit does not handle anything yet but should be responsible for
// text insertion (TODO).
func (i *InputElement) OnTextEdit(str string, start int32, length int32) bool {
	if i.Events.OnTextEdit != nil {
		return i.Events.OnTextEdit(str, start, length)
	}
	return true
}

// HandleUpdate is the method for handling update messages.
func (i *InputElement) HandleUpdate(update UpdateI) {
	switch u := update.(type) {
	case UpdateValue:
		i.SetValue(u.Value)
		i.composition = []rune(u.Value)
		i.cursor = len(i.composition)
	default:
		i.BaseElement.HandleUpdate(update)
	}
}
