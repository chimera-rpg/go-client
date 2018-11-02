package UI

const (
  ORIGIN_RIGHT = 1 << iota
  ORIGIN_BOTTOM
  ORIGIN_CENTERX
  ORIGIN_CENTERY
)

type Number struct {
  Percentage bool
  Value float64
}

func (s *Number) PercentOf(n float64) float64 {
  return n * (s.Value / 100)
}

type Color struct {
  R uint8
  G uint8
  B uint8
  A uint8
}

type Style struct {
  Origin uint8
  X Number
  Y Number
  W Number
  H Number
  PaddingLeft     Number
  PaddingRight    Number
  PaddingTop      Number
  PaddingBottom   Number
  ForegroundColor Color
  BackgroundColor Color
}
