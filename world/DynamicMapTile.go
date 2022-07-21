package world

// DynamicMapTile represents a tile.
type DynamicMapTile struct {
	objects    []*Object
	brightness float32
}

// RemoveObject removes the given object from the tile.
func (d *DynamicMapTile) RemoveObject(o *Object) {
	for i, v := range d.objects {
		if v == o {
			d.objects = append(d.objects[:i], d.objects[i+1:]...)
			return
		}
	}
}

// AddObject adds the given object from the tile.
func (d *DynamicMapTile) AddObject(o *Object) {
	for _, v := range d.objects {
		if v == o {
			return
		}
	}
	d.objects = append(d.objects, o)
}
