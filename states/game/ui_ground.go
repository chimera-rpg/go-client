package game

import "github.com/chimera-rpg/go-client/ui"

type GroundModeWindow struct {
	game         *Game
	container    ui.Container
	toggleButton ui.ElementI
}

func (g *GroundModeWindow) Setup(s *Game) (err error) {
	g.game = s
	g.container.Setup(ui.ContainerConfig{
		Value: "Ground",
		Style: GroundWindowStyle,
	})
	g.toggleButton = ui.NewButtonElement(ui.ButtonElementConfig{
		Value: `nearby`,
		Style: `
			X 0
			Y 0
			W 64
			MinH 20
		`,
		NoFocus: true,
		Events: ui.Events{
			OnMouseButtonUp: func(button uint8, x, y int32) bool {
				var mode GroundMode
				if s.GroundMode == GroundModeNearby {
					mode = GroundModeExact
				} else {
					mode = GroundModeNearby
				}
				s.inputChan <- GroundModeEvent{
					Mode: mode,
				}
				return false
			},
		},
	})
	g.container.GetAdoptChannel() <- g.toggleButton
	s.GameContainer.AdoptChannel <- g.container.This

	return
}

func (g *GroundModeWindow) SyncMode() {
	g.toggleButton.GetUpdateChannel() <- ui.UpdateValue{Value: g.game.GroundMode.String()}
}
