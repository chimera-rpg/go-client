package game

import (
	"fmt"
	"image/color"
	"time"

	"github.com/chimera-rpg/go-client/ui"
	"github.com/chimera-rpg/go-common/network"
)

// Message is a container for a received network message.
type Message struct {
	Received time.Time
	Message  network.CommandMessage
}

// MapMessage represents a floating message on the map.
type MapMessage struct {
	objectID    uint32
	trackObject bool
	x, y, z     int
	el          ui.ElementI
	destroyTime time.Time
	floatY      float64
}

func (s *Game) createMapMessage(y, x, z uint32, body string, col color.RGBA) (MapMessage, error) {
	// Get our initial render position
	xPos, yPos, _ := s.GetRenderPosition(s.world.GetCurrentMap(), y, x, z)

	// Average characters in a word: 4.7; assume slow reading speed 100 wpm, so 1.6 wps; let's assume 4 chars per word so 6 chars per second.
	charsPerSecond := len(body) / 6
	// Ensure minimum of 2 seconds on screen.
	if charsPerSecond < 2 {
		charsPerSecond = 2
	}

	// Create our MapMessage.
	m := MapMessage{
		x: xPos,
		y: yPos,
		el: ui.NewTextElement(ui.TextElementConfig{
			Style: fmt.Sprintf(`
				X %d
				Y %d
				Origin CenterX
				ForegroundColor %d %d %d %d
				OutlineColor 0 0 0 128
				ZIndex 999999
			`, xPos, yPos, col.R, col.G, col.B, col.A),
			Value: body,
		}),
		destroyTime: time.Now().Add(time.Second * time.Duration(charsPerSecond)),
	}

	return m, nil
}
func (s *Game) createMapObjectMessage(objectID uint32, body string, col color.RGBA) (MapMessage, error) {
	o := s.world.GetObject(objectID)
	var x, y, z uint32

	if o != nil {
		x = o.X
		y = o.Y + uint32(o.H) + 1
		z = o.Z
	}

	m, err := s.createMapMessage(y, x, z, body, col)
	if err != nil {
		return m, err
	}
	m.objectID = objectID
	return m, nil
}
