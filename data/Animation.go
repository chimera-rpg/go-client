package data

// Animation provides an AnimationID and FaceID->Frames pairing.
type Animation struct {
	AnimationID uint32
	RandomFrame bool
	Ready       bool
	Faces       []Face
}

func (a Animation) GetFace(id uint32) []AnimationFrame {
	for _, face := range a.Faces {
		if face.FaceID == id {
			return face.Frames
		}
	}
	return make([]AnimationFrame, 0)
}

type Face struct {
	FaceID uint32
	Frames []AnimationFrame
}

// AnimationFrame provides an ImageID and Time pairing.
type AnimationFrame struct {
	ImageID uint32
	Time    int
	X, Y    int8
}
