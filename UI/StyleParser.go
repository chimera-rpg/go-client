package UI

import (
	"github.com/eczarny/lexer"
	"strconv"
	"strings"
)

type styleParser struct {
	lexer        *lexer.Lexer
	currentToken lexer.Token
}

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
	case "PaddingLeft":
		s.PaddingLeft = parseNumber(p.tokenValue())
	case "PaddingRight":
		s.PaddingRight = parseNumber(p.tokenValue())
	case "PaddingTop":
		s.PaddingTop = parseNumber(p.tokenValue())
	case "PaddingBottom":
		s.PaddingBottom = parseNumber(p.tokenValue())
	case "ForegroundColor":
		s.ForegroundColor = parseColor(p.tokenValue())
	case "BackgroundColor":
		s.BackgroundColor = parseColor(p.tokenValue())
	case "Origin":
		s.Origin = parseOrigin(p.tokenValue())
	case "ContentOrigin":
		s.ContentOrigin = parseOrigin(p.tokenValue())
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

func parseColor(s string) (c Color) {
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

func parseOrigin(s string) (f Flags) {
	parts := strings.Split(s, " ")
	for _, n := range parts {
		switch n {
		case "CenterX":
			f.Set(CENTERX)
		case "CenterY":
			f.Set(CENTERY)
		case "Bottom":
			f.Set(BOTTOM)
		case "Right":
			f.Set(RIGHT)
		}
	}
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
