package ui

// BaseElement is our base implementation of the ElementI interface. Every
// Element type must have BaseElement as an anonymous field and override
// any core functionality that it wishes to implement itself.
type BaseElement struct {
	/* NOTE: I'm sure this is The Wrong Way(tm), but we're using an Element interface to point to "this" element via This. This "This" must be set by any BaseElement embedding structs to point to itself via pointer.

	The reason for this is that by using embedded structs to gain a set of default properties and methods for any of the embedding structs, we _must_ either define Adopt/Disown methods on each Element or use a pointer to an interface property that is set in the BaseElement.
	*/
	This      ElementI
	Parent    ElementI
	Children  []ElementI
	Style     Style
	LastStyle Style
	Events    Events
	//
	AdoptChannel   chan ElementI
	DisownChannel  chan ElementI
	DestroyChannel chan bool
	UpdateChannel  chan UpdateI
	// Dirty should be set whenever the Element should be re-rendered
	Dirty bool
	//
	Value     string
	Hidden    bool
	Focusable bool
	Focused   bool
	Holdable  bool
	Held      bool
	// Context is cached when the object is created.
	Context *Context
	// x, y, w, h are cached values from CalculateStyle
	x  int32
	y  int32
	w  int32
	h  int32
	pt int32
	pb int32
	pl int32
	pr int32
	mt int32
	mb int32
	ml int32
	mr int32
}

// Destroy is our stub for destroying an element.
func (b *BaseElement) Destroy() {
	if b.Parent != nil {
		b.Parent.DisownChild(b.This)
	}

	for _, child := range b.Children {
		child.Destroy()
	}
}

// Render handled the rendering of all children and the clearing of the Dirty flag.
// Inheriting elements will generally call this as a super once their own
// rendering is complete.
func (b *BaseElement) Render() {
	for _, child := range b.Children {
		child.Render()
	}
	b.Dirty = false
}

// GetX gets the cached x value.
func (b *BaseElement) GetX() int32 {
	return b.x
}

// GetY gets the cached y value.
func (b *BaseElement) GetY() int32 {
	return b.y
}

// GetWidth gets the cached width value.
func (b *BaseElement) GetWidth() int32 {
	return b.w
}

// GetHeight gets the cached height value.
func (b *BaseElement) GetHeight() int32 {
	return b.h
}

// SetValue sets the text value of the element.
func (b *BaseElement) SetValue(value string) error {
	b.Value = value
	return nil
}

// GetValue retrieves the text value of the element.
func (b *BaseElement) GetValue() string {
	return b.Value
}

// GetStyle returns a pointer to the element's Style.
func (b *BaseElement) GetStyle() *Style {
	return &b.Style
}

// Hit detects if the passed x and y arguments fall within the element's box
func (b *BaseElement) Hit(x int32, y int32) bool {
	if b.Hidden {
		return false
	}
	if b.Parent != nil {
		lx, ly := b.Parent.GetX()+b.x, b.Parent.GetY()+b.y
		if x >= lx && y >= ly && x <= lx+b.w && y <= ly+b.h {
			return true
		}
	} else {
		if x >= b.x && y >= b.y && x <= b.x+b.w && y <= b.y+b.h {
			return true
		}
	}
	return false
}

// CalculateStyle is a heavy method for updating and caching various properties
// for rendering.
func (b *BaseElement) CalculateStyle() {
	if b.Hidden {
		return
	}
	var x, y, w, minw, maxw, h, minh, maxh, pt, pb, pl, pr, mt, mb, ml, mr int32 = b.x, b.y, b.w, 0, 0, b.h, 0, 0, b.pt, b.pb, b.pl, b.pr, b.mt, b.mb, b.ml, b.mr
	if b.Parent != nil {
		if b.Style.X.Percentage {
			x = int32(b.Style.X.PercentOf(float64(b.Parent.GetWidth())))
		} else {
			x = int32(b.Style.X.Value)
		}
		if b.Style.Origin.Has(RIGHT) {
			x = b.Parent.GetWidth() - x
		}
		if b.Parent.IsContainer() {
			//x = x + int32(b.Parent.GetX())
		}
		if b.Style.Y.Percentage {
			y = int32(b.Style.Y.PercentOf(float64(b.Parent.GetHeight())))
		} else {
			y = int32(b.Style.Y.Value)
		}
		if b.Parent.IsContainer() {
			//y = y + int32(b.Parent.GetY())
		}
		if b.Style.Origin.Has(BOTTOM) {
			y = b.Parent.GetHeight() - y
		}
		if b.Style.W.Percentage {
			w = int32(b.Style.W.PercentOf(float64(b.Parent.GetWidth())))
		} else {
			w = int32(b.Style.W.Value)
		}
		if b.Style.H.Percentage {
			h = int32(b.Style.H.PercentOf(float64(b.Parent.GetHeight())))
		} else {
			h = int32(b.Style.H.Value)
		}
		if b.Style.MinW.Percentage {
			minw = int32(b.Style.MinW.PercentOf(float64(b.Parent.GetWidth())))
		} else {
			minw = int32(b.Style.MinW.Value)
		}
		if b.Style.MaxW.Percentage {
			maxw = int32(b.Style.MaxW.PercentOf(float64(b.Parent.GetWidth())))
		} else {
			maxw = int32(b.Style.MaxW.Value)
		}
		if b.Style.MinH.Percentage {
			minh = int32(b.Style.MinH.PercentOf(float64(b.Parent.GetHeight())))
		} else {
			minh = int32(b.Style.MinH.Value)
		}
		if b.Style.MaxH.Percentage {
			maxh = int32(b.Style.MaxH.PercentOf(float64(b.Parent.GetHeight())))
		} else {
			maxh = int32(b.Style.MaxH.Value)
		}

		// Padding
		if b.Style.PaddingLeft.Percentage {
			pl = int32(b.Style.PaddingLeft.PercentOf(float64(b.Parent.GetWidth())))
		} else {
			pl = int32(b.Style.PaddingLeft.Value)
		}
		if b.Style.PaddingRight.Percentage {
			pr = int32(b.Style.PaddingRight.PercentOf(float64(b.Parent.GetWidth())))
		} else {
			pr = int32(b.Style.PaddingRight.Value)
		}
		if b.Style.PaddingTop.Percentage {
			pt = int32(b.Style.PaddingTop.PercentOf(float64(b.Parent.GetHeight())))
		} else {
			pt = int32(b.Style.PaddingTop.Value)
		}
		if b.Style.PaddingBottom.Percentage {
			pb = int32(b.Style.PaddingBottom.PercentOf(float64(b.Parent.GetHeight())))
		} else {
			pb = int32(b.Style.PaddingBottom.Value)
		}
		// Margin
		if b.Style.MarginLeft.Percentage {
			ml = int32(b.Style.MarginLeft.PercentOf(float64(b.Parent.GetWidth())))
		} else {
			ml = int32(b.Style.MarginLeft.Value)
		}
		if b.Style.MarginRight.Percentage {
			mr = int32(b.Style.MarginRight.PercentOf(float64(b.Parent.GetWidth())))
		} else {
			mr = int32(b.Style.MarginRight.Value)
		}
		if b.Style.MarginTop.Percentage {
			mt = int32(b.Style.MarginTop.PercentOf(float64(b.Parent.GetHeight())))
		} else {
			mt = int32(b.Style.MarginTop.Value)
		}
		if b.Style.MarginBottom.Percentage {
			mb = int32(b.Style.MarginBottom.PercentOf(float64(b.Parent.GetHeight())))
		} else {
			mb = int32(b.Style.MarginBottom.Value)
		}

	} else {
		if !b.Style.X.Percentage {
			x = int32(b.Style.X.Value)
		}
		if !b.Style.Y.Percentage {
			y = int32(b.Style.Y.Value)
		}
		if !b.Style.W.Percentage {
			w = int32(b.Style.W.Value)
		}
		if !b.Style.H.Percentage {
			h = int32(b.Style.H.Value)
		}
		if !b.Style.MinW.Percentage {
			minw = int32(b.Style.MinW.Value)
		}
		if !b.Style.MaxW.Percentage {
			maxw = int32(b.Style.MaxW.Value)
		}
		if !b.Style.MinH.Percentage {
			minh = int32(b.Style.MinH.Value)
		}
		if !b.Style.MaxH.Percentage {
			maxh = int32(b.Style.MaxH.Value)
		}
		// Padding
		if !b.Style.PaddingLeft.Percentage {
			pl = int32(b.Style.PaddingLeft.Value)
		}
		if !b.Style.PaddingRight.Percentage {
			pr = int32(b.Style.PaddingRight.Value)
		}
		if !b.Style.PaddingTop.Percentage {
			pt = int32(b.Style.PaddingTop.Value)
		}
		if !b.Style.PaddingBottom.Percentage {
			pb = int32(b.Style.PaddingBottom.Value)
		}
		// Margin
		if !b.Style.MarginLeft.Percentage {
			ml = int32(b.Style.MarginLeft.Value)
		}
		if !b.Style.MarginRight.Percentage {
			mr = int32(b.Style.MarginRight.Value)
		}
		if !b.Style.MarginTop.Percentage {
			mt = int32(b.Style.MarginTop.Value)
		}
		if !b.Style.MarginBottom.Percentage {
			mb = int32(b.Style.MarginBottom.Value)
		}
	}
	if h < minh {
		h = minh
	}
	if maxw > 0 && w > maxw {
		w = maxw
	}
	if w < minw {
		w = minw
	}
	if maxh > 0 && h > maxh {
		h = maxh
	}
	if x != b.x || y != b.y || w != b.w || h != b.h || pl != b.pl || pr != b.pr || pt != b.pt || pb != b.pb || ml != b.ml || mr != b.mr || mt != b.mt || mb != b.mb {
		b.x = x
		b.y = y
		b.w = w + pl + pr
		b.h = h + pt + pb
		b.pl = pl
		b.pr = pr
		b.pt = pt
		b.pb = pb
		b.ml = ml
		b.mr = mr
		b.mt = mt
		b.mb = mb
		b.Dirty = true
	}
	if b.Dirty || b.LastStyle != b.Style {
		if b.Style.Origin.Has(CENTERX) {
			b.x = b.x - b.w/2
		} else if b.Style.Origin.Has(RIGHT) {
			b.x = b.x - b.w - b.mr
		} else {
			b.x = b.x + b.ml
		}
		if b.Style.Origin.Has(CENTERY) {
			b.y = b.y - b.h/2
		} else if b.Style.Origin.Has(BOTTOM) {
			b.y = b.y - b.h - b.mb
		} else {
			b.y = b.y + b.mt
		}
		b.LastStyle = b.Style
		b.Dirty = true
	}
	for _, child := range b.Children {
		child.CalculateStyle()
	}
}

// SetDirty sets the element's dirty flag.
func (b *BaseElement) SetDirty(v bool) {
	b.Dirty = v
}

// IsDirty returns if the dirty flag is set.
func (b *BaseElement) IsDirty() bool {
	return b.Dirty
}

// HasDirt returns if the element is dirty or if any of its children are.
func (b *BaseElement) HasDirt() (dirt bool) {
	dirt = b.IsDirty()
	if dirt == true {
		return
	}
	for _, child := range b.Children {
		dirt = child.HasDirt()
		if dirt {
			return
		}
	}
	return
}

// GetContext returns the rendering context of the element.
func (b *BaseElement) GetContext() *Context {
	return b.Context
}

// SetContext sets the rendering context of the element.
func (b *BaseElement) SetContext(c *Context) {
	b.Context = c
}

// SetParent sets the element's parent to a given Element interface.
func (b *BaseElement) SetParent(e ElementI) {
	if b.Parent != nil && e != nil {
		b.Parent.DisownChild(b.This)
	}
	b.Parent = e
}

// GetParent returns the Element interface parent.
func (b *BaseElement) GetParent() (e ElementI) {
	return b.Parent
}

// DisownChild disowns a given Element interface child.
func (b *BaseElement) DisownChild(c ElementI) {
	for i, child := range b.Children {
		if child == c {
			b.Children = append(b.Children[:i], b.Children[i+1:]...)
			c.SetParent(nil)
			return
		}
	}
}

// AdoptChild adopts a given Element interface as a child.
func (b *BaseElement) AdoptChild(c ElementI) {
	b.Children = append(b.Children, c)
	c.OnAdopted(b)

	// Recalculate our style after adopting.
	b.CalculateStyle()
	b.SetDirty(true)
}

// SetHidden sets the Hidden flag to a particular value, signifying if
// rendering should be skipped during the Render call.
func (b *BaseElement) SetHidden(v bool) {
	b.Hidden = v
}

// IsHidden returns if the element is hidden.
func (b *BaseElement) IsHidden() bool {
	return b.Hidden
}

// SetEvents sets the element's Events to the one passed in.
func (b *BaseElement) SetEvents(e Events) {
	b.Events = e
}

// OnCreated is called when the object is first created.
func (b *BaseElement) OnCreated() {
	if b.Events.OnCreated != nil {
		b.Events.OnCreated()
	}
}

// OnTouchBegin handles touch begin events.
func (b *BaseElement) OnTouchBegin(id uint32, x int32, y int32) bool {
	if b.Events.OnTouchBegin != nil {
		return b.Events.OnTouchBegin(id, x, y)
	}
	return true
}

// OnTouchMove handles touch move events.
func (b *BaseElement) OnTouchMove(id uint32, x int32, y int32) bool {
	if b.Events.OnTouchMove != nil {
		return b.Events.OnTouchMove(id, x, y)
	}
	return true
}

// OnTouchEnd handles touch end events.
func (b *BaseElement) OnTouchEnd(id uint32, x int32, y int32) bool {
	if b.Events.OnTouchEnd != nil {
		return b.Events.OnTouchEnd(id, x, y)
	}
	return true
}

// OnMouseButtonDown handles when a mouse's button is pressed.
func (b *BaseElement) OnMouseButtonDown(buttonID uint8, x int32, y int32) bool {
	if b.Events.OnMouseButtonDown != nil {
		return b.Events.OnMouseButtonDown(buttonID, x, y)
	}
	return true
}

// OnMouseMove handles when a mouse is moved.
func (b *BaseElement) OnMouseMove(x int32, y int32) bool {
	if b.Events.OnMouseMove != nil {
		return b.Events.OnMouseMove(x, y)
	}
	return true
}

// OnMouseIn handles when a mouse enters into the Element.
func (b *BaseElement) OnMouseIn(x int32, y int32) bool {
	if b.Events.OnMouseIn != nil {
		return b.Events.OnMouseIn(x, y)
	}
	return true
}

// OnMouseOut handles when a mouse leaves the Element.
func (b *BaseElement) OnMouseOut(x int32, y int32) bool {
	if b.Events.OnMouseOut != nil {
		return b.Events.OnMouseOut(x, y)
	}
	return true
}

// OnMouseButtonUp handles when a mouse's button is released.
func (b *BaseElement) OnMouseButtonUp(buttonID uint8, x int32, y int32) bool {
	if b.Events.OnMouseButtonUp != nil {
		return b.Events.OnMouseButtonUp(buttonID, x, y)
	}
	return true
}

// OnKeyDown handles when a key is depresed.
func (b *BaseElement) OnKeyDown(key uint8, modifiers uint16, repeat bool) bool {
	if b.Events.OnKeyDown != nil {
		return b.Events.OnKeyDown(key, modifiers, repeat)
	}
	return true
}

// OnKeyUp handles when a key is released.
func (b *BaseElement) OnKeyUp(key uint8, modifiers uint16) bool {
	if b.Events.OnKeyUp != nil {
		return b.Events.OnKeyUp(key, modifiers)
	}
	return true
}

// OnTextInput handles when a text input event is received.
func (b *BaseElement) OnTextInput(str string) bool {
	if b.Events.OnTextInput != nil {
		return b.Events.OnTextInput(str)
	}
	return true
}

// OnTextEdit handles when a text edit event is received.
func (b *BaseElement) OnTextEdit(str string, start int32, length int32) bool {
	if b.Events.OnTextEdit != nil {
		return b.Events.OnTextEdit(str, start, length)
	}
	return true
}

// OnChange is called when the object's value is changed.
func (b *BaseElement) OnChange() {
	if b.Events.OnChange != nil {
		b.Events.OnChange()
	}
}

// OnAdopted is called when an Element is adopted.
func (b *BaseElement) OnAdopted(parent ElementI) {
	b.SetContext(parent.GetContext())
	b.SetParent(parent)
	b.CalculateStyle()
	b.SetDirty(true)
	if b.Events.OnAdopted != nil {
		b.Events.OnAdopted(parent)
	}
}

// CanFocus returns if the element is focusable.
func (b *BaseElement) CanFocus() bool {
	return b.Focusable
}

// SetFocused sets if the element can be focused.
func (b *BaseElement) SetFocused(v bool) {
	b.Focused = v
}

// Focus globally sets the element as the currently focused element.
func (b *BaseElement) Focus() {
	GlobalInstance.FocusElement(b.This)
}

// OnFocus is called when the element is focused.
func (b *BaseElement) OnFocus() bool {
	b.Dirty = true
	if b.Events.OnFocus != nil {
		return b.Events.OnFocus()
	}
	return true
}

// Blur globally blurs the element if it is currently focused.
func (b *BaseElement) Blur() {
	if GlobalInstance.FocusedElement == b.This {
		GlobalInstance.BlurFocusedElement()
	}
}

// OnBlur is called when the element is blurred.
func (b *BaseElement) OnBlur() bool {
	b.Dirty = true
	if b.Events.OnBlur != nil {
		return b.Events.OnBlur()
	}
	return true
}

// CanHold returns if the element should be considered as holdable.
func (b *BaseElement) CanHold() bool {
	return b.Holdable
}

// SetHeld sets the element to be holdable or not.
func (b *BaseElement) SetHeld(v bool) {
	b.Held = v
	b.SetDirty(true)
}

// IsContainer returns if the Element should be considered as an
// element that containers other elements.
func (b *BaseElement) IsContainer() bool {
	return false
}

// GetChildren returns the ElementI children of the element.
func (b *BaseElement) GetChildren() []ElementI {
	return b.Children
}

// SetupChannels sets up the various communication channels.
func (b *BaseElement) SetupChannels() {
	b.AdoptChannel = make(chan ElementI, 1000)
	b.DisownChannel = make(chan ElementI, 1000)
	b.DestroyChannel = make(chan bool, 1)
	b.UpdateChannel = make(chan UpdateI, 1000)
}

// GetAdoptChannel returns the channel used for adopting new elements.
func (b *BaseElement) GetAdoptChannel() chan ElementI {
	return b.AdoptChannel
}

// GetDisownChannel returns the channel used for adopting new elements.
func (b *BaseElement) GetDisownChannel() chan ElementI {
	return b.DisownChannel
}

// GetDestroyChannel returns the channel used for destroying the element.
func (b *BaseElement) GetDestroyChannel() chan bool {
	return b.DestroyChannel
}

// GetUpdateChannel returns the channel used for updating the element.
func (b *BaseElement) GetUpdateChannel() chan UpdateI {
	return b.UpdateChannel
}

// HandleUpdate is the base stub for handling update messages.
func (b *BaseElement) HandleUpdate(update UpdateI) {
	switch u := update.(type) {
	case UpdateValue:
		b.SetValue(u.Value)
	}
}
