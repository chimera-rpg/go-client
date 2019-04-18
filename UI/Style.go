package UI

const (
	ORIGIN_RIGHT = 1 << iota
	ORIGIN_BOTTOM
	ORIGIN_CENTERX
	ORIGIN_CENTERY
)
const (
	CENTERX = 1 << iota
	CENTERY
)

type Number struct {
	IsSet      bool
	Percentage bool
	Value      float64
}

func (s *Number) PercentOf(n float64) float64 {
	return n * (s.Value / 100)
}

func (s *Number) Set(n float64) float64 {
	s.Value = n
	s.IsSet = true
	return s.Value
}

func (s *Number) Unset() {
	s.IsSet = false
}

type Color struct {
	R     uint8
	G     uint8
	B     uint8
	A     uint8
	IsSet bool
}

func (c *Color) Set(r uint8, g uint8, b uint8, a uint8) {
	c.R = r
	c.G = g
	c.B = b
	c.A = a
	c.IsSet = true
}
func (c *Color) Unset() {
	c.IsSet = false
}

type Style struct {
	Origin          uint8
	CenterContent   uint8
	X               Number
	Y               Number
	W               Number
	MinW            Number
	MaxW            Number
	H               Number
	MinH            Number
	MaxH            Number
	PaddingLeft     Number
	PaddingRight    Number
	PaddingTop      Number
	PaddingBottom   Number
	ForegroundColor Color
	BackgroundColor Color
	ResizeToContent bool
}

func (s *Style) Set(o Style) {
	*s = o
	defnum := Number{false, false, 0}
	if s.X != defnum {
		s.X.IsSet = true
	}
	if s.Y != defnum {
		s.Y.IsSet = true
	}
	if s.W != defnum {
		s.W.IsSet = true
	}
	if s.MinW != defnum {
		s.MinW.IsSet = true
	}
	if s.MaxW != defnum {
		s.MaxW.IsSet = true
	}
	if s.H != defnum {
		s.H.IsSet = true
	}
	if s.MinH != defnum {
		s.MinH.IsSet = true
	}
	if s.MaxH != defnum {
		s.MaxH.IsSet = true
	}
	if s.PaddingLeft != defnum {
		s.PaddingLeft.IsSet = true
	}
	if s.PaddingRight != defnum {
		s.PaddingRight.IsSet = true
	}
	if s.PaddingTop != defnum {
		s.PaddingTop.IsSet = true
	}
	if s.PaddingBottom != defnum {
		s.PaddingBottom.IsSet = true
	}
	defcol := Color{0, 0, 0, 0, false}
	if s.ForegroundColor != defcol {
		s.ForegroundColor.IsSet = true
	}
	if s.BackgroundColor != defcol {
		s.BackgroundColor.IsSet = true
	}
	s.ResizeToContent = o.ResizeToContent
}
