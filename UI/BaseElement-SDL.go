// +build !MOBILE
package UI

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

func (b *BaseElement) Destroy() {
}
func (b *BaseElement) Render() {
	for _, child := range b.Children {
		child.Render()
	}
	b.Dirty = false
}
func (b *BaseElement) TouchBegin() {
}
func (b *BaseElement) TouchEnd() {
}
func (b *BaseElement) Pressed(button uint8, state bool, x int, y int) bool {
	return true
}

//
func (b *BaseElement) GetX() int32 {
	return b.x
}
func (b *BaseElement) GetY() int32 {
	return b.y
}
func (b *BaseElement) GetWidth() int32 {
	return b.w
}
func (b *BaseElement) GetHeight() int32 {
	return b.h
}

func (b *BaseElement) SetValue(value string) error {
	b.Value = value
	return nil
}
func (b *BaseElement) GetValue() string {
	return b.Value
}
func (b *BaseElement) GetStyle() *Style {
	return &b.Style
}
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

func (b *BaseElement) SetDirty(v bool) {
	b.Dirty = v
}
func (b *BaseElement) IsDirty() bool {
	return b.Dirty
}
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

func (b *BaseElement) GetContext() *Context {
	return b.Context
}
func (b *BaseElement) SetContext(c *Context) {
	b.Context = c
}

/* Relationships */
func (b *BaseElement) SetParent(e ElementI) {
	if b.Parent != nil && e != nil {
		b.Parent.DisownChild(b.This)
	}
	b.Parent = e
}

func (b *BaseElement) GetParent() (e ElementI) {
	return b.Parent
}

func (b *BaseElement) DisownChild(c ElementI) {
	for i, child := range b.Children {
		if child == c {
			b.Children = append(b.Children[:i], b.Children[i+1:]...)
			c.SetParent(nil)
			return
		}
	}
}

func (b *BaseElement) AdoptChild(c ElementI) {
	b.Children = append(b.Children, c)
	c.OnAdopted(b)
}

func (b *BaseElement) SetHidden(v bool) {
	b.Hidden = v
}
func (b *BaseElement) IsHidden() bool {
	return b.Hidden
}

func (b *BaseElement) SetEvents(e Events) {
	b.Events = e
}

func (b *BaseElement) OnTouchBegin(id uint32, x int32, y int32) bool {
	if b.Events.OnTouchBegin != nil {
		return b.Events.OnTouchBegin(id, x, y)
	}
	return true
}
func (b *BaseElement) OnTouchMove(id uint32, x int32, y int32) bool {
	if b.Events.OnTouchMove != nil {
		return b.Events.OnTouchMove(id, x, y)
	}
	return true
}
func (b *BaseElement) OnTouchEnd(id uint32, x int32, y int32) bool {
	if b.Events.OnTouchEnd != nil {
		return b.Events.OnTouchEnd(id, x, y)
	}
	return true
}
func (b *BaseElement) OnMouseButtonDown(button_id uint8, x int32, y int32) bool {
	if b.Events.OnMouseButtonDown != nil {
		return b.Events.OnMouseButtonDown(button_id, x, y)
	}
	return true
}
func (b *BaseElement) OnMouseMove(x int32, y int32) bool {
	if b.Events.OnMouseMove != nil {
		return b.Events.OnMouseMove(x, y)
	}
	return true
}
func (b *BaseElement) OnMouseIn(x int32, y int32) bool {
	if b.Events.OnMouseIn != nil {
		return b.Events.OnMouseIn(x, y)
	}
	return true
}
func (b *BaseElement) OnMouseOut(x int32, y int32) bool {
	if b.Events.OnMouseOut != nil {
		return b.Events.OnMouseOut(x, y)
	}
	return true
}

func (b *BaseElement) OnMouseButtonUp(button_id uint8, x int32, y int32) bool {
	if b.Events.OnMouseButtonUp != nil {
		return b.Events.OnMouseButtonUp(button_id, x, y)
	}
	return true
}

func (b *BaseElement) OnKeyDown(key uint8, modifiers uint16) bool {
	if b.Events.OnKeyDown != nil {
		return b.Events.OnKeyDown(key, modifiers)
	}
	return true
}

func (b *BaseElement) OnKeyUp(key uint8, modifiers uint16) bool {
	if b.Events.OnKeyUp != nil {
		return b.Events.OnKeyUp(key, modifiers)
	}
	return true
}

func (b *BaseElement) OnTextInput(str string) bool {
	if b.Events.OnTextInput != nil {
		return b.Events.OnTextInput(str)
	}
	return true
}
func (b *BaseElement) OnTextEdit(str string, start int32, length int32) bool {
	if b.Events.OnTextEdit != nil {
		return b.Events.OnTextEdit(str, start, length)
	}
	return true
}

func (b *BaseElement) OnAdopted(parent ElementI) {
	b.SetContext(parent.GetContext())
	b.SetParent(parent)
	b.CalculateStyle()
	b.SetDirty(true)
	if b.Events.OnAdopted != nil {
		b.Events.OnAdopted(parent)
	}
}

func (b *BaseElement) CanFocus() bool {
	return b.Focusable
}
func (b *BaseElement) SetFocused(v bool) {
	b.Focused = v
}

func (b *BaseElement) OnFocus() bool {
	b.Dirty = true
	if b.Events.OnFocus != nil {
		return b.Events.OnFocus()
	}
	return true
}
func (b *BaseElement) OnBlur() bool {
	b.Dirty = true
	if b.Events.OnBlur != nil {
		return b.Events.OnBlur()
	}
	return true
}

func (b *BaseElement) CanHold() bool {
	return b.Holdable
}
func (b *BaseElement) SetHeld(v bool) {
	b.Held = v
	b.SetDirty(true)
}

func (b *BaseElement) IsContainer() bool {
	return false
}

func (b *BaseElement) GetChildren() []ElementI {
	return b.Children
}
