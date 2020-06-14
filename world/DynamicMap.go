package world

// DynamicMap is the dynamically sized map that contains tiles and current objects.
type DynamicMap struct {
	tiles   map[int32]map[int32]map[int32][]DynamicMapTile
	cameraX int32
	cameraY int32
	cameraZ int32
	cameraW int32 // Width of the camera's viewport
	cameraD int32 // Depth of the camera's viewport
	cameraH int32 // Height of the camera's viewport
}

// SetCamera sets the current camera position to x, y, z
func (d *DynamicMap) SetCamera(x, y, z int32) {
	d.cameraX = x
	d.cameraY = y
	d.cameraZ = z
}

// GetCameraView gets a 3D slice of the current camera view.
func (d *DynamicMap) GetCameraView() [][][][]DynamicMapTile {
	tiles := make([][][][]DynamicMapTile, d.cameraW)
	for x := -d.cameraW / 2; x < d.cameraW/2; x++ {
		tiles[d.cameraX+x] = make([][][]DynamicMapTile, d.cameraD)
		for y := -d.cameraD / 2; y < d.cameraD/2; y++ {
			tiles[d.cameraX+x][d.cameraY+y] = make([][]DynamicMapTile, d.cameraH)
			for z := -d.cameraH / 2; z < d.cameraH/2; z++ {
				tiles[d.cameraX+x][d.cameraY+y][d.cameraZ+z] = d.GetTileStack(x, y, z)
			}
		}
	}
	return tiles
}

// GetTileStack gets the tile stack at X, Y, Z.
func (d *DynamicMap) GetTileStack(x, y, z int32) (tiles []DynamicMapTile) {
	if xmap, ok := d.tiles[x]; ok {
		if ymap, ok := xmap[y]; ok {
			if darray, ok := ymap[y]; ok {
				return darray
			}
		}
	}
	return
}
