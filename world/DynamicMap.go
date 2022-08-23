package world

// DynamicMap is the dynamically sized map that contains tiles and current objects.
type DynamicMap struct {
	tiles                []DynamicMapTile
	height, width, depth int
}

// Init initializes the DynamicMap.
func (d *DynamicMap) Init() {
	d.tiles = make([]DynamicMapTile, d.height*d.width*d.depth)
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
	d.tiles[idx] = DynamicMapTile{
		objects: objects,
	}
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
