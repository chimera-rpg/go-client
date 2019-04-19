package UI

type Bits uint8

const (
	CENTERX Bits = 1 << iota
	CENTERY
	BOTTOM
	RIGHT
)

type Number struct {
	Value      float64
	Percentage bool
}

func (s *Number) PercentOf(n float64) float64 {
	return n * (s.Value / 100)
}

func (s *Number) Set(n float64) float64 {
	s.Value = n
	return s.Value
}

type Flags struct {
	Value Bits
}

func (f *Flags) Set(flags Bits) Bits {
	f.Value = f.Value | flags
	return f.Value
}
func (f *Flags) Clear(flags Bits) Bits {
	f.Value = f.Value &^ flags
	return f.Value
}
func (f *Flags) Toggle(flags Bits) Bits {
	f.Value = f.Value ^ flags
	return f.Value
}
func (f *Flags) Has(flags Bits) bool {
	return f.Value&flags != 0
}

type Color struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

func (c *Color) Set(r uint8, g uint8, b uint8, a uint8) {
	c.R = r
	c.G = g
	c.B = b
	c.A = a
}

type Style struct {
	Origin          Flags
	ContentOrigin   Flags
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

func (s *Style) Parse(str string) {
	parser := new(styleParser)
	parser.lexer = NewObjectLexer(str)
	parser.Parse(s)
}
