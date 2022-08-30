package world

// DynamicMapTile represents a tile.
type DynamicMapTile struct {
	objects         []*Object
	brightness      float32
	hue             float32
	sky             float32
	opaque          bool
	finalBrightness float64
	finalHue        float64
}

func (d *DynamicMapTile) Brightness() float64 {
	return float64(d.brightness)
}

func (d *DynamicMapTile) Sky() float64 {
	return float64(d.sky)
}

func (d *DynamicMapTile) Hue() float64 {
	return float64(d.hue)
}
func (d *DynamicMapTile) FinalBrightness() float64 {
	return d.finalBrightness
}

func (d *DynamicMapTile) FinalHue() float64 {
	return d.finalHue
}

func (d *DynamicMapTile) Objects() []*Object {
	return d.objects
}

func (d *DynamicMapTile) Refresh() {
	for _, o := range d.objects {
		if o.Opaque {
			d.opaque = true
		}
	}
}

// RemoveObject removes the given object from the tile.
func (d *DynamicMapTile) RemoveObject(o *Object) {
	for i, v := range d.objects {
		if v == o {
			d.objects = append(d.objects[:i], d.objects[i+1:]...)
			d.Refresh()
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
	d.Refresh()
}
