package world

// DynamicMap is the dynamically sized map that contains tiles and current objects.
type DynamicMap struct {
	tiles                [][][]DynamicMapTile
	height, width, depth uint32
	cameraX              uint32
	cameraY              uint32
	cameraZ              uint32
	cameraW              uint32 // Width of the camera's viewport
	cameraD              uint32 // Depth of the camera's viewport
	cameraH              uint32 // Height of the camera's viewport
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

// SetCamera sets the current camera position to y, x, z
func (d *DynamicMap) SetCamera(y, x, z uint32) {
	d.cameraY = y
	d.cameraX = x
	d.cameraZ = z
}

// GetCameraView gets a 3D slice of the current camera view.
func (d *DynamicMap) GetCameraView() [][][]DynamicMapTile {
	tile := make([][][]DynamicMapTile, d.cameraW)
	for x := -d.cameraW / 2; x < d.cameraW/2; x++ {
		tile[d.cameraX+x] = make([][]DynamicMapTile, d.cameraD)
		for y := -d.cameraD / 2; y < d.cameraD/2; y++ {
			tile[d.cameraX+x][d.cameraY+y] = make([]DynamicMapTile, d.cameraH)
			for z := -d.cameraH / 2; z < d.cameraH/2; z++ {
				tile[d.cameraX+x][d.cameraY+y][d.cameraZ+z] = d.GetTile(int(y), int(x), int(z))
			}
		}
	}
	return tile
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
