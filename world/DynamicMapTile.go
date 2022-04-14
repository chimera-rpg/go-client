package world

// DynamicMapTile represents a tile.
type DynamicMapTile struct {
	objectIDs  []uint32
	brightness float32
}

// GetObjects returns the contained objectIDs in a tile.
func (d *DynamicMapTile) GetObjects() []uint32 {
	return d.objectIDs
}

// RemoveObject removes the given objectID from the tile.
func (d *DynamicMapTile) RemoveObject(oID uint32) {
	for i, v := range d.objectIDs {
		if v == oID {
			d.objectIDs = append(d.objectIDs[:i], d.objectIDs[i+1:]...)
		}
	}
}
