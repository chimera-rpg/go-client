package ui

import "image/color"

// Bits is the type used for Flags.
type Bits uint8

// These const values are the underlying bit flags used for various
// positioning options.
const (
	CENTERX Bits = 1 << iota
	CENTERY
	BOTTOM
	RIGHT
)

// These const values are the underlying bit flags used for various
// resizing options.
const (
	TOCONTENT Bits = 1 << iota
)

// Number is our special number container type.
type Number struct {
	Value      float64
	Percentage bool
}

// PercentOf returns the percent of the target float that this Number
// would fill.
func (s *Number) PercentOf(n float64) float64 {
	return n * (s.Value / 100)
}

// Set sets the Value to the given float.
func (s *Number) Set(n float64) float64 {
	s.Value = n
	return s.Value
}

// Flags are byte flags used for various Style usage.
type Flags struct {
	Value Bits
}

// Set sets our Flags value to the given bits.
func (f *Flags) Set(flags Bits) Bits {
	f.Value = f.Value | flags
	return f.Value
}

// Clear clears our Flags value.
func (f *Flags) Clear(flags Bits) Bits {
	f.Value = f.Value &^ flags
	return f.Value
}

// Toggle toggles on or off a given flag.
func (f *Flags) Toggle(flags Bits) Bits {
	f.Value = f.Value ^ flags
	return f.Value
}

// Has checks if Flags has some flags set.
func (f *Flags) Has(flags Bits) bool {
	return f.Value&flags != 0
}

// Style the type used by Elements to control desired positioning and styling.
type Style struct {
	Origin          Flags
	ContentOrigin   Flags
	Resize          Flags
	X               Number
	Y               Number
	W               Number
	MinW            Number
	MaxW            Number
	H               Number
	MinH            Number
	MaxH            Number
	ZIndex          Number
	PaddingLeft     Number
	PaddingRight    Number
	PaddingTop      Number
	PaddingBottom   Number
	MarginLeft      Number
	MarginRight     Number
	MarginTop       Number
	MarginBottom    Number
	ScrollLeft      Number
	ScrollTop       Number
	ForegroundColor color.NRGBA
	BackgroundColor color.NRGBA
	OutlineColor    color.NRGBA
}

// Parse parses the given style string into property changes in the given Style.
func (s *Style) Parse(str string) {
	parser := new(styleParser)
	parser.lexer = NewObjectLexer(str)
	parser.Parse(s)
}
