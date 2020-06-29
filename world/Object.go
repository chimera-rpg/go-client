package world

type Object struct {
	ID          uint32
	Type        uint8
	AnimationID uint32
	FaceID      uint32
	Y, X, Z     uint32 // We keep Y, X, Z information here to make it easier to render objects. This is updated when Tile updates are received.
	H, W, D     uint8
	Missing     bool // Represents if the object is currently in an unknown location. This happens when a Tile that holds an Object no longer holds it and no other Tile has claimed it.
}

// ObjectsFilter returns a new slice containing all Objects in the slice that satisfy the predicate f.
func ObjectsFilter(vo []Object, f func(Object) bool) []Object {
	vof := make([]Object, 0)
	for _, v := range vo {
		if f(v) {
			vof = append(vof, v)
		}
	}
	return vof
}
