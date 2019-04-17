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
	Value  string
	Hidden bool
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
	if x >= b.x && y >= b.y && x <= b.x+b.w && y <= b.y+b.h {
		return true
	}
	return false
}
func (b *BaseElement) CalculateStyle() {
	if b.Hidden {
		return
	}
	var x, y, w, h, pt, pb, pl, pr int32 = b.x, b.y, b.w, b.h, b.pt, b.pb, b.pl, b.pr
	if b.Parent != nil {
		if b.Style.X.IsSet {
			if b.Style.X.Percentage {
				x = int32(b.Style.X.PercentOf(float64(b.Parent.GetWidth())))
			} else {
				x = int32(b.Style.X.Value)
			}
			x = x + b.Parent.GetX()
		}
		if b.Style.Y.IsSet {
			if b.Style.Y.Percentage {
				y = int32(b.Style.Y.PercentOf(float64(b.Parent.GetHeight())))
			} else {
				y = int32(b.Style.Y.Value)
			}
			y = y + int32(b.Parent.GetY())
		}
		if b.Style.W.IsSet {
			if b.Style.W.Percentage {
				w = int32(b.Style.W.PercentOf(float64(b.Parent.GetWidth())))
			} else {
				w = int32(b.Style.W.Value)
			}
		}
		if b.Style.H.IsSet {
			if b.Style.H.Percentage {
				h = int32(b.Style.H.PercentOf(float64(b.Parent.GetHeight())))
			} else {
				h = int32(b.Style.H.Value)
			}
		}
		// Padding
		if b.Style.PaddingLeft.IsSet {
			if b.Style.PaddingLeft.Percentage {
				pl = int32(b.Style.PaddingLeft.PercentOf(float64(b.Parent.GetWidth())))
			} else {
				pl = int32(b.Style.PaddingLeft.Value)
			}
		}
		if b.Style.PaddingRight.IsSet {
			if b.Style.PaddingRight.Percentage {
				pr = int32(b.Style.PaddingRight.PercentOf(float64(b.Parent.GetWidth())))
			} else {
				pr = int32(b.Style.PaddingRight.Value)
			}
		}
		if b.Style.PaddingTop.IsSet {
			if b.Style.PaddingTop.Percentage {
				pt = int32(b.Style.PaddingTop.PercentOf(float64(b.Parent.GetHeight())))
			} else {
				pt = int32(b.Style.PaddingTop.Value)
			}
		}
		if b.Style.PaddingBottom.IsSet {
			if b.Style.PaddingBottom.Percentage {
				pb = int32(b.Style.PaddingBottom.PercentOf(float64(b.Parent.GetHeight())))
			} else {
				pb = int32(b.Style.PaddingBottom.Value)
			}
		}
	} else {
		if b.Style.X.IsSet && !b.Style.X.Percentage {
			x = int32(b.Style.X.Value)
		}
		if b.Style.Y.IsSet && !b.Style.Y.Percentage {
			y = int32(b.Style.Y.Value)
		}
		if b.Style.W.IsSet && !b.Style.W.Percentage {
			w = int32(b.Style.W.Value)
		}
		if b.Style.H.IsSet && !b.Style.H.Percentage {
			h = int32(b.Style.H.Value)
		}
		// Padding
		if b.Style.PaddingLeft.IsSet && !b.Style.PaddingLeft.Percentage {
			pl = int32(b.Style.PaddingLeft.Value)
		}
		if b.Style.PaddingRight.IsSet && !b.Style.PaddingRight.Percentage {
			pr = int32(b.Style.PaddingRight.Value)
		}
		if b.Style.PaddingTop.IsSet && !b.Style.PaddingTop.Percentage {
			pt = int32(b.Style.PaddingTop.Value)
		}
		if b.Style.PaddingBottom.IsSet && !b.Style.PaddingBottom.Percentage {
			pb = int32(b.Style.PaddingBottom.Value)
		}
	}
	if x != b.x || y != b.y || w != b.w || h != b.h || pl != b.pl || pr != b.pr || pt != b.pt || pb != b.pb {
		b.x = x
		b.y = y
		b.w = w + pl + pr
		b.h = h + pt + pb
		b.pl = pl
		b.pr = pr
		b.pt = pt
		b.pb = pb
		b.Dirty = true
	}
	if b.Dirty || b.LastStyle != b.Style {
		if b.Style.Origin&ORIGIN_CENTERX != 0 {
			b.x = b.x - b.w/2
		} else if b.Style.Origin&ORIGIN_RIGHT != 0 {
			b.x = b.x - b.w
		}
		if b.Style.Origin&ORIGIN_CENTERY != 0 {
			b.y = b.y - b.h/2
		} else if b.Style.Origin&ORIGIN_BOTTOM != 0 {
			b.y = b.y - b.h
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
func (b *BaseElement) OnMouseButtonUp(button_id uint8, x int32, y int32) bool {
	if b.Events.OnMouseButtonUp != nil {
		return b.Events.OnMouseButtonUp(button_id, x, y)
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

func (b *BaseElement) GetChildren() []ElementI {
	return b.Children
}
