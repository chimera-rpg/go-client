package data

// Animation provides an AnimationID and FaceID->Frames pairing.
type Animation struct {
	AnimationID uint32
	RandomFrame bool
	Ready       bool
	Faces       map[uint32][]AnimationFrame
}

func (a Animation) GetFace(id uint32) []AnimationFrame {
	if frames, ok := a.Faces[id]; ok {
		return frames
	}
	return make([]AnimationFrame, 0)
}

// AnimationFrame provides an ImageID and Time pairing.
type AnimationFrame struct {
	ImageID uint32
	Time    int
	X, Y    int8
}
