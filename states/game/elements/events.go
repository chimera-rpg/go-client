package elements

// MouseInput is the UserInput for mouse events.
type MouseInput struct {
	X, Y     int32
	Button   uint8
	Pressed  bool
	Held     bool
	Released bool
}

type MouseMoveInput struct {
	X, Y int32
}
