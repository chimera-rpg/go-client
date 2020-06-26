package world

type Object struct {
	ID          uint32
	Type        uint8
	AnimationID uint32
	FaceID      uint32
	Y, X, Z     uint32 // We keep Y, X, Z information here to make it easier to render objects. This is updated when Tile updates are received.
	Gone        bool   // Represents if the object is currently in an unknown location. This happens when a Tile that holds an Object no longer holds it and no other Tile has claimed it.
}
