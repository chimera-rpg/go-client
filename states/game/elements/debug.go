package elements

import (
	"fmt"

	"github.com/chimera-rpg/go-client/ui"
)

type DebugWindow struct {
	game               game
	show               bool
	container          *ui.Container
	worldInfo          ui.ElementI
	tileInfo           ui.ElementI
	tileLightInfo      ui.ElementI
	selfLightInfo      ui.ElementI
	underfootLightInfo ui.ElementI
}

func (c *DebugWindow) Setup(game game, style string, inputChan chan interface{}) (*ui.Container, error) {
	c.game = game
	var err error
	c.container, err = ui.NewContainerElement(ui.ContainerConfig{
		Value: "Debug",
		Style: style,
	})
	if err != nil {
		return nil, err
	}
	c.worldInfo = ui.NewTextElement(ui.TextElementConfig{
		Value: "",
		Style: `
			ForegroundColor 255 255 255 255
			OutlineColor 0 0 0 255
		`,
	})
	c.tileInfo = ui.NewTextElement(ui.TextElementConfig{
		Value: "",
		Style: `
			Y 12
			ForegroundColor 255 255 255 255
			OutlineColor 0 0 0 255
		`,
	})
	c.tileLightInfo = ui.NewTextElement(ui.TextElementConfig{
		Value: "",
		Style: `
			Y 24
			ForegroundColor 255 255 255 255
			OutlineColor 0 0 0 255
		`,
	})
	c.selfLightInfo = ui.NewTextElement(ui.TextElementConfig{
		Value: "",
		Style: `
			Y 36
			ForegroundColor 255 255 255 255
			OutlineColor 0 0 0 255
		`,
	})
	c.underfootLightInfo = ui.NewTextElement(ui.TextElementConfig{
		Value: "",
		Style: `
			Y 48
			ForegroundColor 255 255 255 255
			OutlineColor 0 0 0 255
		`,
	})

	c.container.GetAdoptChannel() <- c.worldInfo
	c.container.GetAdoptChannel() <- c.tileInfo
	c.container.GetAdoptChannel() <- c.tileLightInfo
	c.container.GetAdoptChannel() <- c.selfLightInfo
	c.container.GetAdoptChannel() <- c.underfootLightInfo

	return c.container, nil
}

// Refresh refreshes the window to show focused, change images, etc.
func (c *DebugWindow) Refresh() {
	if !c.show {
		return
	}
	if m := c.game.World().GetCurrentMap(); m != nil {
		c.worldInfo.GetUpdateChannel() <- ui.UpdateValue{Value: fmt.Sprintf("%dx%dx%d, outdoor %t, outdoor brightness %f, ambient brightness %f, ambient hue %f", m.GetHeight(), m.GetWidth(), m.GetDepth(), m.Outdoor(), m.OutdoorBrightness(), m.AmbientBrightness(), m.AmbientHue())}
		if vo := c.game.World().GetViewObject(); vo != nil {
			c.selfLightInfo.GetUpdateChannel() <- ui.UpdateValue{Value: fmt.Sprintf("self: brightness %f, hue %f", vo.Brightness, vo.Hue)}
			if t := m.GetTile(vo.Y, vo.X, vo.Z); t != nil {
				c.tileInfo.GetUpdateChannel() <- ui.UpdateValue{Value: fmt.Sprintf("%dx%dx%d: %d objects", vo.Y, vo.X, vo.Z, len(t.Objects()))}
				c.tileLightInfo.GetUpdateChannel() <- ui.UpdateValue{Value: fmt.Sprintf("brightness %f, sky %f, finalBrightness %f, finalHue %f", t.Brightness(), t.Sky(), t.FinalBrightness(), t.FinalHue())}
			}
			if t := m.GetTile(vo.Y-1, vo.X, vo.Z); t != nil {
				c.underfootLightInfo.GetUpdateChannel() <- ui.UpdateValue{Value: fmt.Sprintf("underfoot: brightness %f(%f), hue %f(%f)\n", t.FinalBrightness(), t.Brightness(), t.FinalHue(), t.Hue())}
			}
		}
	}
}

func (c *DebugWindow) Toggle() {
	c.show = !c.show
	c.container.GetUpdateChannel() <- ui.UpdateHidden(!c.show)
	c.Refresh()
}
