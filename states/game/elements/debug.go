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
	blockedInfo        ui.ElementI
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
	c.blockedInfo = ui.NewTextElement(ui.TextElementConfig{
		Value: "",
		Style: `
			Y 60
			ForegroundColor 255 255 255 255
			OutlineColor 0 0 0 255
		`,
	})

	c.container.GetAdoptChannel() <- c.worldInfo
	c.container.GetAdoptChannel() <- c.tileInfo
	c.container.GetAdoptChannel() <- c.tileLightInfo
	c.container.GetAdoptChannel() <- c.selfLightInfo
	c.container.GetAdoptChannel() <- c.underfootLightInfo
	c.container.GetAdoptChannel() <- c.blockedInfo

	return c.container, nil
}

// Refresh refreshes the window to show focused, change images, etc.
func (c *DebugWindow) Refresh() {
	if !c.show {
		return
	}
	if m := c.game.World().GetCurrentMap(); m != nil {
		or, og, ob := m.OutdoorRGB()
		ar, ag, ab := m.AmbientRGB()
		c.worldInfo.GetUpdateChannel() <- ui.UpdateValue{Value: fmt.Sprintf("%dx%dx%d, outdoor %t, outdoor rgb %d:%d:%d, ambient rgb %d:%d:%d", m.GetHeight(), m.GetWidth(), m.GetDepth(), m.Outdoor(), or, og, ob, ar, ag, ab)}
		if vo := c.game.World().GetViewObject(); vo != nil {
			c.selfLightInfo.GetUpdateChannel() <- ui.UpdateValue{Value: fmt.Sprintf("self rgb %d:%d:%d", vo.R, vo.G, vo.B)}
			if t := m.GetTile(vo.Y, vo.X, vo.Z); t != nil {
				c.tileInfo.GetUpdateChannel() <- ui.UpdateValue{Value: fmt.Sprintf("%dx%dx%d: %d objects", vo.Y, vo.X, vo.Z, len(t.Objects()))}
				tr, tg, tb := t.RGB()
				fr, fg, fb := t.RGB()
				c.tileLightInfo.GetUpdateChannel() <- ui.UpdateValue{Value: fmt.Sprintf("rgb %d:%d:%d, sky %f, final rgb %d:%d:%d", tr, tg, tb, t.Sky(), fr, fg, fb)}
			}
			if t := m.GetTile(vo.Y-1, vo.X, vo.Z); t != nil {
				tr, tg, tb := t.RGB()
				fr, fg, fb := t.RGB()
				c.underfootLightInfo.GetUpdateChannel() <- ui.UpdateValue{Value: fmt.Sprintf("underfoot: rgb %d:%d:%d(%d:%d:%d)\n", tr, tg, tb, fr, fg, fb)}
			}
		}
	}
	c.blockedInfo.GetUpdateChannel() <- ui.UpdateValue{Value: fmt.Sprintf("room %t, left open %t, above open %t, front open %t", c.game.World().InRoom, !c.game.World().LeftBlocked, !c.game.World().AboveBlocked, !c.game.World().FrontBlocked)}
}

func (c *DebugWindow) Toggle() {
	c.show = !c.show
	c.container.GetUpdateChannel() <- ui.UpdateHidden(!c.show)
	c.Refresh()
}
