package world

// DynamicMapTile represents a tile.
type DynamicMapTile struct {
	objectIDs []uint32
}

func (d *DynamicMapTile) GetObjects() []uint32 {
	return d.objectIDs
}
