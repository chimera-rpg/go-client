package ui

import (
	"fmt"
	"strconv"
	"strings"

	"image/color"

	"github.com/eczarny/lexer"
)

type styleParser struct {
	lexer        *lexer.Lexer
	currentToken lexer.Token
}

// ParseStyle parses the given string into the passed Style.
func ParseStyle(s *Style, str string) {
	parser := new(styleParser)
	parser.lexer = NewObjectLexer(str)
	parser.Parse(s)
	return
}

func (p *styleParser) Parse(s *Style) {
Loop:
	for {
		switch p.nextToken().Type {
		case TokenProperty:
			p.parseProperty(s, p.tokenValue())
		case TokenEOF:
			break Loop
		default:
			p.nextToken()
		}
	}
}

func (p *styleParser) parseProperty(s *Style, prop string) {
	p.nextToken()
	switch prop {
	case "X":
		s.X = parseNumber(p.tokenValue())
	case "Y":
		s.Y = parseNumber(p.tokenValue())
	case "W":
		s.W = parseNumber(p.tokenValue())
	case "MinW":
		s.MinW = parseNumber(p.tokenValue())
	case "MaxW":
		s.MaxW = parseNumber(p.tokenValue())
	case "H":
		s.H = parseNumber(p.tokenValue())
	case "MinH":
		s.MinH = parseNumber(p.tokenValue())
	case "MaxH":
		s.MaxH = parseNumber(p.tokenValue())
	case "ZIndex":
		s.ZIndex = parseNumber(p.tokenValue())
	case "Scale":
		parts := strings.Split(p.tokenValue(), " ")
		partsLen := len(parts)
		if partsLen == 1 { // xy
			s.ScaleX, s.ScaleY = parseNumber(parts[0]), parseNumber(parts[0])
		} else if partsLen == 2 { // x y
			s.ScaleX, s.ScaleY = parseNumber(parts[0]), parseNumber(parts[1])
		}
	case "ScaleX":
		s.ScaleX = parseNumber(p.tokenValue())
	case "ScaleY":
		s.ScaleY = parseNumber(p.tokenValue())
	case "Margin":
		parts := strings.Split(p.tokenValue(), " ")
		partsLen := len(parts)
		if partsLen == 1 { // ltrb%
			s.MarginRight, s.MarginLeft, s.MarginTop, s.MarginBottom = parseNumber(parts[0]), parseNumber(parts[0]), parseNumber(parts[0]), parseNumber(parts[0])
		} else if partsLen == 2 { // lr% tb%
			s.MarginLeft, s.MarginRight = parseNumber(parts[0]), parseNumber(parts[0])
			s.MarginTop, s.MarginBottom = parseNumber(parts[1]), parseNumber(parts[1])
		} else if partsLen == 3 { // l% t% r%
			s.MarginLeft = parseNumber(parts[0])
			s.MarginTop = parseNumber(parts[1])
			s.MarginRight = parseNumber(parts[2])
		} else if partsLen == 4 { // l% t% r% b%
			s.MarginLeft = parseNumber(parts[0])
			s.MarginTop = parseNumber(parts[1])
			s.MarginRight = parseNumber(parts[2])
			s.MarginBottom = parseNumber(parts[3])
		}
	case "MarginLeft":
		s.MarginLeft = parseNumber(p.tokenValue())
	case "MarginRight":
		s.MarginRight = parseNumber(p.tokenValue())
	case "MarginTop":
		s.MarginTop = parseNumber(p.tokenValue())
	case "MarginBottom":
		s.MarginBottom = parseNumber(p.tokenValue())
	case "Padding":
		parts := strings.Split(p.tokenValue(), " ")
		partsLen := len(parts)
		if partsLen == 1 { // ltrb%
			s.PaddingRight, s.PaddingLeft, s.PaddingTop, s.PaddingBottom = parseNumber(parts[0]), parseNumber(parts[0]), parseNumber(parts[0]), parseNumber(parts[0])
		} else if partsLen == 2 { // lr% tb%
			s.PaddingLeft, s.PaddingRight = parseNumber(parts[0]), parseNumber(parts[0])
			s.PaddingTop, s.PaddingBottom = parseNumber(parts[1]), parseNumber(parts[1])
		} else if partsLen == 3 { // l% t% r%
			s.PaddingLeft = parseNumber(parts[0])
			s.PaddingTop = parseNumber(parts[1])
			s.PaddingRight = parseNumber(parts[2])
		} else if partsLen == 4 { // l% t% r% b%
			s.PaddingLeft = parseNumber(parts[0])
			s.PaddingTop = parseNumber(parts[1])
			s.PaddingRight = parseNumber(parts[2])
			s.PaddingBottom = parseNumber(parts[3])
		}
	case "PaddingLeft":
		s.PaddingLeft = parseNumber(p.tokenValue())
	case "PaddingRight":
		s.PaddingRight = parseNumber(p.tokenValue())
	case "PaddingTop":
		s.PaddingTop = parseNumber(p.tokenValue())
	case "PaddingBottom":
		s.PaddingBottom = parseNumber(p.tokenValue())
	case "Scroll":
		parts := strings.Split(p.tokenValue(), " ")
		partsLen := len(parts)
		if partsLen == 1 { // l
			s.ScrollLeft = parseNumber(parts[0])
		} else if partsLen == 2 { // l t
			s.ScrollLeft = parseNumber(parts[0])
			s.ScrollTop = parseNumber(parts[1])
		}
	case "ScrollLeft":
		s.ScrollLeft = parseNumber(p.tokenValue())
	case "ScrollTop":
		s.ScrollTop = parseNumber(p.tokenValue())
	case "ForegroundColor":
		s.ForegroundColor = parseColor(p.tokenValue())
	case "BackgroundColor":
		s.BackgroundColor = parseColor(p.tokenValue())
	case "OutlineColor":
		s.OutlineColor = parseColor(p.tokenValue())
	case "ScrollbarGripperColor":
		s.ScrollbarGripperColor = parseColor(p.tokenValue())
	case "Origin":
		s.Origin = parseOrigin(s, p.tokenValue())
	case "ContentOrigin":
		s.ContentOrigin = parseOrigin(s, p.tokenValue())
	case "Resize":
		s.Resize = parseResize(p.tokenValue())
	case "Display":
		s.Display = parseDisplay(s, p.tokenValue())
	case "Direction":
		s.Direction = parseDirection(s, p.tokenValue())
	case "Wrap":
		s.Wrap = parseWrap(s, p.tokenValue())
	case "Overflow":
		s.Overflow = parseOverflow(s, p.tokenValue())
	case "Alpha":
		s.Alpha = parseNumber(p.tokenValue())
	case "ColorMod":
		s.ColorMod = parseColor(p.tokenValue())
	}
}

func parseNumber(s string) (n Number) {
	if s[len(s)-1:] == "%" {
		n.Percentage = true
		v, _ := strconv.ParseFloat(s[:len(s)-1], 64)
		n.Value = v
	} else {
		v, _ := strconv.ParseFloat(s, 64)
		n.Value = v
	}
	return
}

func parseColor(s string) (c color.NRGBA) {
	parts := strings.Split(s, " ")
	for i, n := range parts {
		v, _ := strconv.ParseUint(n, 10, 8)
		switch i {
		case 0: // r
			c.R = uint8(v)
		case 1: // g
			c.G = uint8(v)
		case 2: // b
			c.B = uint8(v)
		case 3: // a
			c.A = uint8(v)
		}
	}
	return
}

func parseResize(s string) (f Flags) {
	parts := strings.Split(s, " ")
	for _, n := range parts {
		switch n {
		case "ToContent":
			f.Set(TOCONTENT)
		}
	}
	return
}

func parseOrigin(style *Style, s string) (f Flags) {
	f = style.Origin
	parts := strings.Split(s, " ")
	for _, n := range parts {
		switch n {
		case "CenterX":
			f.Set(CENTERX)
			f.Clear(RIGHT)
		case "CenterY":
			f.Set(CENTERY)
			f.Clear(BOTTOM)
		case "Bottom":
			f.Set(BOTTOM)
			f.Clear(CENTERY)
		case "Right":
			f.Set(RIGHT)
			f.Clear(CENTERX)
		case "Top":
			f.Clear(CENTERY)
			f.Clear(BOTTOM)
		case "Left":
			f.Clear(CENTERX)
			f.Clear(RIGHT)
		}
	}
	return
}

func parseDisplay(style *Style, s string) (f Flags) {
	f = style.Display
	switch s {
	case "Columns":
		f.Set(COLUMNS)
		f.Clear(ROWS)
	case "Rows":
		f.Set(ROWS)
		f.Clear(COLUMNS)
	}
	return
}

func parseDirection(style *Style, s string) (f Flags) {
	f = style.Direction
	switch s {
	case "Regular":
		f.Set(REGULAR)
		f.Clear(REVERSE)
	case "Reverse":
		f.Set(REVERSE)
		f.Clear(REGULAR)
	}
	return
}

func parseWrap(style *Style, s string) (f Flags) {
	f = style.Wrap
	switch s {
	case "NoWrap":
		f.Set(NOWRAP)
		f.Clear(WRAP)
	case "Wrap":
		f.Set(WRAP)
		f.Clear(NOWRAP)
	case "Hard":
		f.Set(WRAP)
		f.Set(HARD)
		f.Clear(NOWRAP)
	}
	return
}

func parseOverflow(style *Style, s string) (f Flags) {
	f = style.Overflow
	switch s {
	case "X":
		f.Set(OVERFLOWX)
	case "Y":
		f.Set(OVERFLOWY)
	}
	fmt.Println(f)
	return
}

//

func (p *styleParser) tokenValue() string {
	return p.currentToken.Value.(string)
}

func (p *styleParser) nextToken() lexer.Token {
	p.currentToken = p.lexer.NextToken()
	return p.currentToken
}

func (p *styleParser) acceptToken(tokenType lexer.TokenType) bool {
	return p.nextToken().Type == tokenType
}

func (p *styleParser) expectToken(tokenType lexer.TokenType, v interface{}) string {
	if !p.acceptToken(tokenType) {
		panic(v)
	}
	return p.tokenValue()
}
