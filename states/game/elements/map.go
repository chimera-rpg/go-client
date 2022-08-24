package elements

import (
	"math"
	"time"

	"github.com/chimera-rpg/go-client/ui"
)

type MapWindow struct {
	Container ui.Container
	Messages  []MapMessage
}

// MapMessage represents a floating message on the map.
type MapMessage struct {
	ObjectID    uint32
	TrackObject bool
	X, Y, Z     int
	El          ui.ElementI
	DestroyTime time.Time
	FloatY      float64
}

func (m *MapWindow) Setup(style string, inputChan chan interface{}) (ui.Container, error) {
	// Sub-window: map
	err := m.Container.Setup(ui.ContainerConfig{
		Style: style,
		Events: ui.Events{
			OnMouseButtonDown: func(buttonID uint8, x int32, y int32) bool {
				inputChan <- MouseInput{
					Button:  buttonID,
					Pressed: false,
					X:       x,
					Y:       y,
				}
				return true
			},
			OnMouseButtonUp: func(buttonID uint8, x int32, y int32) bool {
				inputChan <- MouseInput{
					Button:  buttonID,
					Pressed: true,
					X:       x,
					Y:       y,
				}
				return true
			},
			OnMouseMove: func(x, y int32) bool {
				inputChan <- MouseMoveInput{
					X: x,
					Y: y,
				}
				return true
			},
			OnHold: func(buttonID uint8, x, y int32) bool {
				inputChan <- MouseInput{
					Button:  buttonID,
					Pressed: true,
					Held:    true,
					X:       x,
					Y:       y,
				}
				return true
			},
			OnUnhold: func(buttonID uint8, x, y int32) bool {
				inputChan <- MouseInput{
					Button:   buttonID,
					Pressed:  true,
					Released: true,
					X:        x,
					Y:        y,
				}
				return true
			},
		},
	})

	return m.Container, err
}

func (m *MapWindow) MouseAngleFromView(x, y int32) float64 {
	x1 := x - m.Container.GetAbsoluteX()
	y1 := y - m.Container.GetAbsoluteY()
	x2 := m.Container.GetWidth() / 2
	y2 := m.Container.GetHeight() / 2
	dY := y2 - y1
	dX := x2 - x1
	dA := (math.Atan2(float64(dY), float64(dX)) * 180 / math.Pi) + 180
	return dA

}
