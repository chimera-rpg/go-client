package states

import (
	"fmt"
	"image/color"
	"time"

	"github.com/chimera-rpg/go-client/ui"
)

// MapMessage represents a floating message on the map.
type MapMessage struct {
	objectID    uint32
	x, y, z     int
	el          ui.ElementI
	destroyTime time.Time
}

func (s *Game) createMapMessage(objectID uint32, body string, col color.RGBA) (MapMessage, error) {
	o := s.world.GetObject(objectID)
	var x, y, z uint32

	if o != nil {
		x = o.X
		y = o.Y + uint32(o.H) + 1
		z = o.Z
	}
	// Get our initial render position
	xPos, yPos, _ := s.GetRenderPosition(s.world.GetCurrentMap(), y, x, z)

	// Create our MapMessage.
	m := MapMessage{
		objectID: objectID,
		x:        xPos,
		y:        yPos,
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
		destroyTime: time.Now().Add(time.Millisecond * time.Duration(200*len(body))),
	}

	return m, nil
}
