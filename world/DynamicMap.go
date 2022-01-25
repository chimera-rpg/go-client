package world

// DynamicMap is the dynamically sized map that contains tiles and current objects.
type DynamicMap struct {
	tiles                [][][]DynamicMapTile
	height, width, depth uint32
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
func (d *DynamicMap) SetTile(y, x, z uint32, objectIDs []uint32) {
	if int(y) >= len(d.tiles) || int(x) >= len(d.tiles[0]) || int(z) >= len(d.tiles[0][0]) {
		return
	}
	d.tiles[y][x][z] = DynamicMapTile{objectIDs}
}

// GetTile gets the tile stack at Y, X, Z.
func (d *DynamicMap) GetTile(y, x, z int) (tiles DynamicMapTile) {
	if int(y) >= len(d.tiles) || int(x) >= len(d.tiles[0]) || int(z) >= len(d.tiles[0][0]) {
		return
	}
	return d.tiles[y][x][z]
}

// GetHeight gets height.
func (d *DynamicMap) GetHeight() uint32 {
	return d.height
}

// GetWidth gets width.
func (d *DynamicMap) GetWidth() uint32 {
	return d.width
}

// GetDepth gets depth.
func (d *DynamicMap) GetDepth() uint32 {
	return d.depth
}
