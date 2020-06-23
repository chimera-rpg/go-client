package data

type Animation struct {
	AnimationID uint32
	Faces       map[uint32][]AnimationFrame
}

type AnimationFrame struct {
	ImageID uint32
	Time    int
}
