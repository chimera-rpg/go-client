package ui

// Instance is the managing instance of the entire UI system.
type Instance struct {
	dataManager     DataManagerI
	HeldElement     ElementI
	FocusedElement  ElementI
	HoveredElements []ElementI
	Running         bool
	RootWindow      Window
	Context         Context
}

// GlobalInstance is our pointer to the GlobalInstance. Used for Focus/Blur
// calls from within Elements.
var GlobalInstance *Instance

// CheckChannels handles iterating through all element channels.
func (instance *Instance) CheckChannels(e ElementI) {
	var ok, valid bool
	// Destruction checking
	select {
	case <-e.GetDestroyChannel():
		e.Destroy()
		return
	default:
		break
	}
	// Adoption checking
	for {
		var adoption ElementI
		select {
		case adoption, valid = <-e.GetAdoptChannel():
			ok = true
		default:
			ok = false
		}
		if ok && valid {
			//fmt.Printf("Got child: %v\n", adoption)
			e.AdoptChild(adoption)
		} else if !ok {
			break
		}
	}
	// Disown checking
	for {
		var disowned ElementI
		select {
		case disowned, valid = <-e.GetDisownChannel():
			ok = true
		default:
			ok = false
		}
		if ok && valid {
			e.DisownChild(disowned)
		} else if !ok {
			break
		}
	}
	// Update checking
	for {
		var update UpdateI
		select {
		case update, valid = <-e.GetUpdateChannel():
			ok = true
		default:
			ok = false
		}
		if ok && valid {
			e.HandleUpdate(update)
		} else if !ok {
			break
		}
	}
	for _, child := range e.GetChildren() {
		instance.CheckChannels(child)
	}

}

// BlurFocusedElement blurs the current focused element if it exists.
func (instance *Instance) BlurFocusedElement() {
	if instance.FocusedElement != nil {
		instance.FocusedElement.SetFocused(false)
		instance.FocusedElement.OnBlur()
	}
	instance.FocusedElement = nil
}

// FocusElement focuses the target element, blurring the previous element if
// it exists.
func (instance *Instance) FocusElement(e ElementI) {
	if instance.FocusedElement != nil && instance.FocusedElement != e {
		instance.FocusedElement.SetFocused(false)
		instance.FocusedElement.OnBlur()
	}
	e.SetFocused(true)
	e.OnFocus()
	instance.FocusedElement = e
}

// FocusNextElement finds and focuses the next focusable element after
// the passed element.
func (instance *Instance) FocusNextElement(start ElementI) {
	found := false
	for _, c := range start.GetParent().GetChildren() {
		if c == start {
			found = true
		} else if found {
			if c.CanFocus() {
				instance.FocusElement(c)
				return
			}
		}
	}
	// if we get here just Blur the focused one
	instance.BlurFocusedElement()
}

// FocusPreviousElement finds and focuses the previous element before
// the passed element.
func (instance *Instance) FocusPreviousElement(start ElementI) {
	found := false
	children := start.GetParent().GetChildren()
	for i := len(children) - 1; i >= 0; i-- {
		c := children[i]
		if c == start {
			found = true
		} else if found {
			if c.CanFocus() {
				instance.FocusElement(c)
				return
			}
		}
	}
	// if we get here just Blur the focused one
	instance.BlurFocusedElement()
}
