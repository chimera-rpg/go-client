package ui

import (
	"image/color"
	"sort"
)

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
	Dirty   bool
	Restyle bool
	//
	Value     string
	Hidden    bool
	Focusable bool
	Focused   bool
	Holdable  bool
	Held      bool
	OOB       bool
	// Context is cached when the object is created.
	Context *Context
	// x, y, w, h are cached values from CalculateStyle
	x int32
	y int32
	w int32
	h int32
	// ax, ay are cached absolute values.
	ax int32
	ay int32
	pt int32
	pb int32
	pl int32
	pr int32
	mt int32
	mb int32
	ml int32
	mr int32
	// scroll
	sl int32
	st int32
}

// BaseElementConfig provides teh configuration for a new BaseElement.
type BaseElementConfig struct {
	Style  string
	Events Events
}

// NewBaseElement creates a new BaseElement using the passed configuration.
func NewBaseElement(c BaseElementConfig) ElementI {
	b := BaseElement{}
	b.This = ElementI(&b)
	b.Style.Parse(c.Style)
	b.Events = c.Events
	b.SetupChannels()
	b.OnCreated()

	return ElementI(&b)
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
	// Sort by ZIndex before rendering. FIXME: This should only be resorted when Z indices actually change.
	sort.Slice(b.Children, func(i, j int) bool {
		return b.Children[i].GetZIndex() < b.Children[j].GetZIndex()
	})
	// Render.
	for _, child := range b.VisibleChildren() {
		child.Render()
	}
	b.RenderPost()
	b.Dirty = false
}

// RenderPost is a special rendering that is called after all the elements in a container have been rendered.
func (b *BaseElement) RenderPost() {
}

// SetX gets the cached x value.
func (b *BaseElement) SetX(x int32) {
	b.ax += b.x - x
	b.x = x
}

// GetX gets the cached x value.
func (b *BaseElement) GetX() int32 {
	return b.x
}

// SetY sets the cached y value.
func (b *BaseElement) SetY(y int32) {
	b.ay += b.y - y
	b.y = y
}

// GetY gets the cached y value.
func (b *BaseElement) GetY() int32 {
	return b.y
}

// GetAbsoluteX gets the cached absolute x value.
func (b *BaseElement) GetAbsoluteX() int32 {
	return b.ax
}

// GetAbsoluteY gets the cached absolute y value.
func (b *BaseElement) GetAbsoluteY() int32 {
	return b.ay
}

// GetWidth gets the cached width value.
func (b *BaseElement) GetWidth() int32 {
	return b.w
}

// GetHeight gets the cached height value.
func (b *BaseElement) GetHeight() int32 {
	return b.h
}

// GetScrollLeft gets the cached scroll left value.
func (b *BaseElement) GetScrollLeft() int32 {
	return b.sl
}

// GetScrollTop gets the cached scroll top value.
func (b *BaseElement) GetScrollTop() int32 {
	return b.st
}

// GetZIndex returns the element's rendering index.
func (b *BaseElement) GetZIndex() int {
	return int(b.Style.ZIndex.Value)
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

// Hit detects if the passed x and y arguments fall within the element's absolute box
func (b *BaseElement) Hit(x int32, y int32) bool {
	if b.IsHidden() {
		return false
	}
	if x >= b.ax && y >= b.ay && x <= b.ax+b.w && y <= b.ay+b.h {
		return true
	}
	return false
}

// PixelHit detects if a pixel-perfect hit is made. Only usable with Image, all others use Hit to resolve.
func (b *BaseElement) PixelHit(x int32, y int32) bool {
	if b.IsHidden() {
		return false
	}
	return b.Hit(x, y)
}

// CalculateStyle is a heavy method for updating and caching various properties
// for rendering.
func (b *BaseElement) CalculateStyle() {
	if b.IsHidden() {
		return
	}
	var x, y, ax, ay, w, minw, maxw, h, minh, maxh, pt, pb, pl, pr, mt, mb, ml, mr, sl, st int32 = b.x, b.y, b.ax, b.ay, b.w, 0, 0, b.h, 0, 0, b.pt, b.pb, b.pl, b.pr, b.mt, b.mb, b.ml, b.mr, b.sl, b.st
	if b.Parent != nil {
		if b.Style.X.Percentage {
			x = int32(b.Style.X.PercentOf(float64(b.Parent.GetWidth())))
		} else {
			x = int32(b.Style.X.Value)
		}
		if !b.Parent.IsContainer() {
			ax = int32(b.Parent.GetAbsoluteX()) + x
			x = int32(b.Parent.GetX()) + x
		} else {
			ax = int32(b.Parent.GetAbsoluteX()) + x
		}
		x -= b.Parent.GetScrollLeft()
		ax -= b.Parent.GetScrollLeft()
		if b.Style.Origin.Has(RIGHT) {
			x = b.Parent.GetWidth() - x
			ax = b.Parent.GetAbsoluteX() + b.Parent.GetWidth() - ax
		}
		var relY int32
		if b.Style.Y.Percentage {
			relY = int32(b.Style.Y.PercentOf(float64(b.Parent.GetHeight())))
		} else {
			relY = int32(b.Style.Y.Value)
		}
		if !b.Parent.IsContainer() {
			y = int32(b.Parent.GetY()) + relY
		} else {
			y = relY
		}
		ay = int32(b.Parent.GetAbsoluteY()) + relY
		y -= b.Parent.GetScrollTop()
		ay -= b.Parent.GetScrollTop()
		if b.Style.Origin.Has(BOTTOM) {
			y = b.Parent.GetHeight() - relY
			ay += (b.Parent.GetHeight() - relY*2) // W...why  do we do relY*2
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
		ax = x
		if !b.Style.Y.Percentage {
			y = int32(b.Style.Y.Value)
		}
		ay = y
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

	// Check if we're out of bounds relative to our parent.
	if b.Parent != nil {
		pw := b.Parent.GetWidth() / 2
		ph := b.Parent.GetHeight() / 2
		if x >= -pw && y >= -h && (x-w) <= pw+pw && (y-h) <= ph+ph {
			b.OOB = false
		} else {
			b.OOB = true
		}
	}

	// Scroll
	if b.Style.ScrollLeft.Percentage {
		sl = int32(b.Style.ScrollLeft.PercentOf(float64(w)))
	} else {
		sl = int32(b.Style.ScrollLeft.Value)
	}
	if b.Style.ScrollTop.Percentage {
		st = int32(b.Style.ScrollTop.PercentOf(float64(h)))
	} else {
		st = int32(b.Style.ScrollTop.Value)
	}

	if x != b.x || y != b.y || ax != b.ax || ay != b.ay || w != b.w || h != b.h || pl != b.pl || pr != b.pr || pt != b.pt || pb != b.pb || ml != b.ml || mr != b.mr || mt != b.mt || mb != b.mb || sl != b.sl || st != b.st {
		b.x = x
		b.y = y
		b.ax = ax
		b.ay = ay
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
		b.st = st
		b.sl = sl
		b.Dirty = true
	}
	if b.Dirty || b.LastStyle != b.Style {
		if b.Style.Origin.Has(CENTERX) {
			b.x = b.x - b.w/2
			b.ax = b.ax - b.w/2
		} else if b.Style.Origin.Has(RIGHT) {
			b.x = b.x - b.w - b.mr
			b.ax = b.ax - b.w - b.mr
		} else {
			b.x = b.x + b.ml
			b.ax = b.ax + b.ml
		}
		if b.Style.Origin.Has(CENTERY) {
			b.y = b.y - b.h/2
			b.ay = b.ay - b.h/2
		} else if b.Style.Origin.Has(BOTTOM) {
			b.y = b.y - b.h - b.mb
			b.ay = b.ay - b.h - b.mb
		} else {
			b.y = b.y + b.mt
			b.ay = b.ay + b.mt
		}
		b.LastStyle = b.Style
		b.Dirty = true
		for _, child := range b.Children {
			child.CalculateStyle()
		}
	}
	b.Restyle = false
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

func (b *BaseElement) ShouldRestyle() bool {
	return b.Restyle
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
	c.OnAdopted(b.This)

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
	/*if b.GetParent() != nil {
		if b.GetParent().IsHidden() {
			return true
		}
	}*/
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

// OnPressed handles when a mouse's button is released.
func (b *BaseElement) OnPressed(buttonID uint8, x int32, y int32) bool {
	if b.Events.OnPressed != nil {
		return b.Events.OnPressed(buttonID, x, y)
	}
	return true
}

// OnHold handles when a button press is held for a duration.
func (b *BaseElement) OnHold(buttonID uint8, x int32, y int32) bool {
	if b.Events.OnHold != nil {
		return b.Events.OnHold(buttonID, x, y)
	}
	return true
}

// OnUnhold handles when a button is released after held for a duration.
func (b *BaseElement) OnUnhold(buttonID uint8, x int32, y int32) bool {
	if b.Events.OnUnhold != nil {
		return b.Events.OnUnhold(buttonID, x, y)
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

// OnTextSubmit handles when a text edit event is received.
func (b *BaseElement) OnTextSubmit(str string) bool {
	if b.Events.OnTextSubmit != nil {
		return b.Events.OnTextSubmit(str)
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

// OnWindowResized is called when an element in the element's heirarchy has been resized.
func (b *BaseElement) OnWindowResized(w, h int32) {
	if b.Events.OnWindowResized != nil {
		b.Events.OnWindowResized(w, h)
	}
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
	dirty := true
	switch u := update.(type) {
	case UpdateValue:
		b.SetValue(u.Value)
	case UpdateX:
		b.Style.X = u.Number
		b.Restyle = true
	case UpdateY:
		b.Style.Y = u.Number
		b.Restyle = true
	case UpdateW:
		b.Style.W = u.Number
		b.Restyle = true
	case UpdateH:
		b.Style.H = u.Number
		b.Restyle = true
	case UpdateDimensions:
		b.Style.X.Value = u.X.Value
		b.Style.Y.Value = u.Y.Value
		b.Style.W.Value = u.W.Value
		b.Style.H.Value = u.H.Value
		b.Restyle = true
	case UpdateScroll:
		b.Style.ScrollLeft = u.Left
		b.Style.ScrollTop = u.Top
		b.Restyle = true
	case UpdateScrollLeft:
		b.Style.ScrollLeft = u.Number
		b.Restyle = true
	case UpdateScrollTop:
		b.Style.ScrollTop = u.Number
		b.Restyle = true
	case UpdateZIndex:
		b.Style.ZIndex = u.Number
	case UpdateOutlineColor:
		b.Style.OutlineColor = u
	case UpdateBackgroundColor:
		b.Style.BackgroundColor = color.NRGBA(u)
	case UpdateForegroundColor:
		b.Style.ForegroundColor = color.NRGBA(u)
	case UpdateDirt:
		dirty = u
	case UpdateFocus:
		b.Focus()
	case UpdateHidden:
		b.SetHidden(bool(u))
	case UpdateAlpha:
		b.Style.Alpha.Set(u)
	case UpdateColorMod:
		b.Style.ColorMod = color.NRGBA{u.R, u.G, u.B, u.A}
	}
	b.SetDirty(dirty)
}

func (b *BaseElement) IsGrayscale() bool {
	return false
}

func (b *BaseElement) IsOOB() bool {
	return b.OOB
}

func (b *BaseElement) VisibleChildren() (vchilds []ElementI) {
	for _, c := range b.Children {
		if !c.IsOOB() {
			vchilds = append(vchilds, c)
		}
	}
	return
}
