package UI

type BaseElement struct {
  Parent ElementI
  Children []ElementI
  Style Style
  LastStyle Style
  // Dirty should be set whenever the Element should be re-rendered
  Dirty bool
  //
  Value string
  // Context is cached when the object is created.
  Context *Context
  // x, y, w, h are cached values from CalculateStyle
  x int32
  y int32
  w int32
  h int32
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
func (b *BaseElement) CalculateStyle() {
  var x, y, w, h, pt, pb, pl, pr int32 = 0, 0, 0, 0, 0, 0, 0, 0
  if b.Parent != nil {
    if b.Style.X.Percentage {
      x = int32(b.Style.X.PercentOf(float64(b.Parent.GetWidth())))
    } else {
      x = int32(b.Style.X.Value)
    }
    if b.Style.Y.Percentage {
      y = int32(b.Style.Y.PercentOf(float64(b.Parent.GetHeight())))
    } else {
      y = int32(b.Style.Y.Value)
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
    if b.Style.Origin & ORIGIN_CENTERX != 0 {
      b.x = b.x - b.w / 2
    } else if b.Style.Origin & ORIGIN_RIGHT != 0 {
      b.x = b.x - b.w
    }
    if b.Style.Origin & ORIGIN_CENTERY != 0 {
      b.y = b.y - b.h / 2
    } else if b.Style.Origin & ORIGIN_BOTTOM != 0 {
      b.y = b.y - b.h
    }
    b.LastStyle = b.Style
    b.Dirty = true
  }
  for _, child := range b.Children {
    child.CalculateStyle()
  }
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
    b.Parent.DisownChild(ElementI(b))
  }
  b.Parent = e
}

func (b *BaseElement) GetParent() (e ElementI) {
  return b.Parent
}

func (b *BaseElement) AdoptChild(e ElementI) {
  e.SetContext(b.Context)
  b.Children = append(b.Children, e)
  e.SetParent(b)
  e.CalculateStyle()
}

func (b *BaseElement) DisownChild(e ElementI) {
  for i, child := range b.Children {
    if child == e {
      b.Children = append(b.Children[:i], b.Children[i+1:]...)
      e.SetParent(nil)
      return
    }
  }
}

func (b *BaseElement) GetChildren() *[]ElementI {
  return &b.Children
}
