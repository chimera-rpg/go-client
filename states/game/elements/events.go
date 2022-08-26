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

// FocusObject
type FocusObjectEvent struct {
	ID uint32
}

// HoverObjectEvent
type HoverObjectEvent struct {
	ID uint32
}

// UnhoverObjectEvent
type UnhoverObjectEvent struct {
	ID uint32
}

// ResizeEvent is used to notify the UI of a resize change.
type ResizeEvent struct{}

// ChatEvent is used to send an input chat to the main loop.
type ChatEvent struct {
	Body string
}

// InspectRequestEvent sends an inspect request for an object.
type InspectRequestEvent struct {
	ID uint32
}
