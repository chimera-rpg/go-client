package world

// DynamicMapTile represents a tile.
type DynamicMapTile struct {
	objectIDs []uint32
}

// GetObjects returns the contained objectIDs in a tile.
func (d *DynamicMapTile) GetObjects() []uint32 {
	return d.objectIDs
}
