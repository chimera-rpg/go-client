package world

// DynamicMap is the dynamically sized map that contains tiles and current objects.
type DynamicMap struct {
	tiles                [][][]DynamicMapTile
	height, width, depth int
}

// Init initializes the DynamicMap.
func (d *DynamicMap) Init() {
	d.tiles = make([][][]DynamicMapTile, d.height)
	for i := range d.tiles {
		d.tiles[i] = make([][]DynamicMapTile, d.width)
		for j := range d.tiles[i] {
			d.tiles[i][j] = make([]DynamicMapTile, d.depth)
		}
	}

}

// SetTile sets the tile at y, x, z
func (d *DynamicMap) SetTile(y, x, z uint32, objects []*Object) {
	if int(y) >= len(d.tiles) || int(x) >= len(d.tiles[0]) || int(z) >= len(d.tiles[0][0]) {
		return
	}
	d.tiles[y][x][z] = DynamicMapTile{
		objects: objects,
	}
}

// GetTile gets the tile stack at Y, X, Z.
func (d *DynamicMap) GetTile(y, x, z int) (tiles *DynamicMapTile) {
	if y < 0 || y >= len(d.tiles) || x < 0 || x >= len(d.tiles[0]) || z < 0 || z >= len(d.tiles[0][0]) {
		return
	}
	return &d.tiles[y][x][z]
}

func (d *DynamicMap) SetTileLight(y, x, z uint32, brightness float32) {
	if int(y) >= len(d.tiles) || int(x) >= len(d.tiles[0]) || int(z) >= len(d.tiles[0][0]) {
		return
	}
	d.tiles[y][x][z].brightness = brightness
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
