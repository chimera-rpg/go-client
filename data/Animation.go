package data

// Animation provides an AnimationID and FaceID->Frames pairing.
type Animation struct {
	AnimationID uint32
	Faces       map[uint32][]AnimationFrame
}

// AnimationFrame provides an ImageID and Time pairing.
type AnimationFrame struct {
	ImageID uint32
	Time    int
	X, Y    int8
}
