package world

// TileKey is a key for our dynamic map hash.
type TileKey struct {
	Y, X, Z uint32
}

// DynamicMap is the dynamically sized map that contains tiles and current objects.
type DynamicMap struct {
	tiles   map[TileKey]DynamicMapTile
	cameraX uint32
	cameraY uint32
	cameraZ uint32
	cameraW uint32 // Width of the camera's viewport
	cameraD uint32 // Depth of the camera's viewport
	cameraH uint32 // Height of the camera's viewport
}

// Init initializes the DynamicMap.
func (d *DynamicMap) Init() {
	d.tiles = make(map[TileKey]DynamicMapTile)
}

// SetTile sets the tile at y, x, z
func (d *DynamicMap) SetTile(y, x, z uint32, objectIDs []uint32) {
	d.tiles[TileKey{y, x, z}] = DynamicMapTile{objectIDs}
}

// SetCamera sets the current camera position to x, y, z
func (d *DynamicMap) SetCamera(x, y, z uint32) {
	d.cameraX = x
	d.cameraY = y
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
				tile[d.cameraX+x][d.cameraY+y][d.cameraZ+z] = d.GetTile(x, y, z)
			}
		}
	}
	return tile
}

// GetTileStack gets the tile stack at X, Y, Z.
func (d *DynamicMap) GetTile(x, y, z uint32) (tiles DynamicMapTile) {
	if darray, ok := d.tiles[TileKey{y, x, z}]; ok {
		return darray
	}
	return
}
