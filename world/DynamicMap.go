package world

import (
	"math"
)

// DynamicMap is the dynamically sized map that contains tiles and current objects.
type DynamicMap struct {
	tiles                         []DynamicMapTile
	unblockedTiles                [][][]bool
	height, width, depth          int
	outdoor                       bool
	outdoorBrightness             float64
	ambientHue, ambientBrightness float64
}

// Init initializes the DynamicMap.
func (d *DynamicMap) Init() {
	d.tiles = make([]DynamicMapTile, d.height*d.width*d.depth)
	d.unblockedTiles = make([][][]bool, 0)
}

func (d *DynamicMap) Index(y, x, z int) int {
	return (d.height*d.width*z + d.width*y) + x
}
func (d *DynamicMap) At(y, x, z int) *DynamicMapTile {
	return &d.tiles[d.Index(y, x, z)]
}

// SetTile sets the tile at y, x, z
func (d *DynamicMap) SetTile(y, x, z int, objects []*Object) {
	if y >= d.height || x >= d.width || z >= d.depth {
		return
	}
	idx := d.Index(y, x, z)
	d.tiles[idx].objects = objects
	d.tiles[idx].Refresh()
}

// GetTile gets the tile stack at Y, X, Z.
func (d *DynamicMap) GetTile(y, x, z int) (tiles *DynamicMapTile) {
	if y < 0 || y >= d.height || x < 0 || x >= d.width || z < 0 || z >= d.depth {
		return
	}
	return &d.tiles[d.Index(y, x, z)]
}

func (d *DynamicMap) SetTileLight(y, x, z int, brightness float32) {
	if y < 0 || y >= d.height || x < 0 || x >= d.width || z < 0 || z >= d.depth {
		return
	}
	d.tiles[d.Index(y, x, z)].brightness = brightness
}

func (d *DynamicMap) SetTileHue(y, x, z int, hue float32) {
	if y < 0 || y >= d.height || x < 0 || x >= d.width || z < 0 || z >= d.depth {
		return
	}
	d.tiles[d.Index(y, x, z)].hue = hue
}

func (d *DynamicMap) SetTileSky(y, x, z int, sky float32) {
	if y < 0 || y >= d.height || x < 0 || x >= d.width || z < 0 || z >= d.depth {
		return
	}
	d.tiles[d.Index(y, x, z)].sky = sky
}

func (d *DynamicMap) RecalculateLightingAt(y, x, z int) {
	index := d.Index(y, x, z)
	d.tiles[index].finalBrightness = d.BrightnessAt(y, x, z)
	d.tiles[index].finalHue = d.HueAt(y, x, z)
	/*return &d.tiles[]
	t := d.At(y, x, z)
	t.finalBrightness = d.BrightnessAt(y, x, z)
	t.finalHue = d.HueAt(y, x, z)*/
}

func (d *DynamicMap) BrightnessAt(y, x, z int) float64 {
	brightness := d.ambientBrightness
	if d.outdoor {
		brightness += d.outdoorBrightness * float64(d.tiles[d.Index(y, x, z)].sky)
	}
	brightness += float64(d.tiles[d.Index(y, x, z)].brightness)

	return math.Max(0, math.Min(brightness, 1.0))
}

func (d *DynamicMap) HueAt(y, x, z int) float64 {
	hue := d.ambientHue
	hue += float64(d.tiles[d.Index(y, x, z)].hue)
	return hue
}

// GetHeight gets height.
func (d *DynamicMap) GetHeight() int {
	return d.height
}

// GetWidth gets width.
func (d *DynamicMap) GetWidth() int {
	return d.width
}

// GetDepth gets depth.
func (d *DynamicMap) GetDepth() int {
	return d.depth
}

func (d *DynamicMap) Outdoor() bool {
	return d.outdoor
}

func (d *DynamicMap) OutdoorBrightness() float64 {
	return d.outdoorBrightness
}

func (d *DynamicMap) AmbientBrightness() float64 {
	return d.ambientBrightness
}

func (d *DynamicMap) AmbientHue() float64 {
	return d.ambientHue
}
