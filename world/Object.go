package world

import (
	"image"
	"time"

	"github.com/chimera-rpg/go-client/ui"
)

// Object represents an arbitrary map object.
type Object struct {
	ID                 uint32
	Type               uint8
	AnimationID        uint32
	FaceID             uint32
	FrameIndex         int           // The current frame index.
	FrameElapsed       time.Duration // The amount of time elapsed for the object's current frame.
	Index              int           // Position in its owning Tile.
	Y, X, Z            uint32        // We keep Y, X, Z information here to make it easier to render objects. This is updated when Tile updates are received.
	H, W, D            uint8
	Missing            bool // Represents if the object is currently in an unknown location. This happens when a Tile that holds an Object no longer holds it and no other Tile has claimed it.
	WasMissing         bool
	Changed            bool        // Represents if the object's position has been changed. Cleared by Game.RenderObject
	Squeezing          bool        // Represents if the object is squeezing. Causes the rendered image to be lightly squashed in the X axis.
	Crouching          bool        // Represents if the object is crouching. Causes the rendered image to be lightly squashed in the Y axis.
	Opaque             bool        // Represents if the object is considered to block vision.
	Visible            bool        // Represents if the object is visible.
	VisibilityChange   bool        // Used to record if the visibility of the object has changed since last render.
	Unblocked          bool        // Represents if the object is unblocked (should be alpha).
	UnblockedChange    bool        // Used to record if the unblocked state of the object has changed since last render.
	LightingChange     bool        // Used to represent if the lighting of the object has changed.
	Brightness         float32     // How much additional brightness should be applied...?
	Element            ui.ElementI // This is kind of bad, but it's simpler for rendering if we pair the ui element with the object directly.
	FrameImageID       uint32
	Image              image.Image // Image reference, also stored here for faster access...
	OutOfVision        bool        // Represents if the object is out of the character's vision.
	OutOfVisionChanged bool        // Represents if the object is out of the character's vision.
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
