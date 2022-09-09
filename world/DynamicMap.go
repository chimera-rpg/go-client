package world

// DynamicMap is the dynamically sized map that contains tiles and current objects.
type DynamicMap struct {
	tiles                                 []DynamicMapTile
	unblockedTiles                        [][][]bool
	height, width, depth                  int
	outdoor                               bool
	outdoorRed, outdoorGreen, outdoorBlue uint8
	ambientRed, ambientGreen, ambientBlue uint8
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

func (d *DynamicMap) SetTileLight(y, x, z int, r, g, b uint8) {
	if y < 0 || y >= d.height || x < 0 || x >= d.width || z < 0 || z >= d.depth {
		return
	}
	t := d.GetTile(y, x, z)
	t.r = r
	t.g = g
	t.b = b
}

func (d *DynamicMap) SetTileSky(y, x, z int, sky float32) {
	if y < 0 || y >= d.height || x < 0 || x >= d.width || z < 0 || z >= d.depth {
		return
	}
	d.tiles[d.Index(y, x, z)].sky = sky
}

func (d *DynamicMap) RecalculateLightingAt(y, x, z int) {
	index := d.Index(y, x, z)

	r := uint16(d.ambientRed)
	g := uint16(d.ambientGreen)
	b := uint16(d.ambientBlue)
	if d.outdoor {
		r += uint16(float64(d.outdoorRed) * float64(d.tiles[d.Index(y, x, z)].sky))
		g += uint16(float64(d.outdoorGreen) * float64(d.tiles[d.Index(y, x, z)].sky))
		b += uint16(float64(d.outdoorBlue) * float64(d.tiles[d.Index(y, x, z)].sky))
	}

	r += uint16(d.tiles[index].r)
	g += uint16(d.tiles[index].g)
	b += uint16(d.tiles[index].b)

	if r >= 255 {
		r = 255
	}
	if g >= 255 {
		g = 255
	}
	if b >= 255 {
		b = 255
	}
	d.tiles[index].finalRed = uint8(r)
	d.tiles[index].finalGreen = uint8(g)
	d.tiles[index].finalBlue = uint8(b)
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

func (d *DynamicMap) OutdoorRGB() (uint8, uint8, uint8) {
	return d.outdoorRed, d.outdoorGreen, d.outdoorBlue
}

func (d *DynamicMap) AmbientRGB() (uint8, uint8, uint8) {
	return d.ambientRed, d.ambientGreen, d.ambientBlue
}
