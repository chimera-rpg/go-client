package world

import (
	"image"
	"time"

	"github.com/chimera-rpg/go-client/data"
	"github.com/chimera-rpg/go-client/ui"
	cdata "github.com/chimera-rpg/go-common/data"
)

// Object represents an arbitrary map object.
type Object struct {
	ID                                                                                             uint32
	Type                                                                                           uint8
	AnimationID                                                                                    uint32
	Animation                                                                                      *data.Animation
	FaceID                                                                                         uint32
	Face                                                                                           data.Face
	Frame                                                                                          *data.AnimationFrame
	FrameIndex                                                                                     int           // The current frame index.
	FrameElapsed                                                                                   time.Duration // The amount of time elapsed for the object's current frame.
	Index                                                                                          int           // Position in its owning Tile.
	Y, X, Z                                                                                        int           // We keep Y, X, Z information here to make it easier to render objects. This is updated when Tile updates are received.
	H, W, D                                                                                        int8
	Reach                                                                                          uint8
	Missing                                                                                        bool // Represents if the object is currently in an unknown location. This happens when a Tile that holds an Object no longer holds it and no other Tile has claimed it.
	WasMissing                                                                                     bool
	Changed                                                                                        bool        // Represents if the object's position has been changed. Cleared by Game.RenderObject
	Squeezing                                                                                      bool        // Represents if the object is squeezing. Causes the rendered image to be lightly squashed in the X axis.
	Crouching                                                                                      bool        // Represents if the object is crouching. Causes the rendered image to be lightly squashed in the Y axis.
	Opaque                                                                                         bool        // Represents if the object is considered to block vision.
	Visible                                                                                        bool        // Represents if the object is visible.
	VisibilityChange                                                                               bool        // Used to record if the visibility of the object has changed since last render.
	Unblocked                                                                                      bool        // Represents if the object is unblocked (should be alpha).
	UnblockedChange                                                                                bool        // Used to record if the unblocked state of the object has changed since last render.
	LightingChange                                                                                 bool        // Used to represent if the lighting of the object has changed.
	Brightness                                                                                     float64     // How much additional brightness should be applied...?
	Hue                                                                                            float64     // Th' hue
	Element                                                                                        ui.ElementI // This is kind of bad, but it's simpler for rendering if we pair the ui element with the object directly.
	HasShadow                                                                                      bool        // Used to indicate that the object should have a shadow. Defined on object creation based upon its archetype type.
	ShadowElement                                                                                  ui.ElementI // This is also kind of bad.
	FrameImageID                                                                                   uint32
	Image                                                                                          image.Image // Image reference, also stored here for faster access...
	RenderX, RenderY, RenderZ                                                                      int         // Cached render positions, set in RenderObject after Changes is set to true.
	Adjusted                                                                                       bool
	AdjustX, AdjustY                                                                               int
	FinalRenderX, FinalRenderOffsetX, FinalRenderY, FinalRenderOffsetY, FinalRenderW, FinalRenderH int  // Represents the _final_ rendering positions, including scaling.
	RecalculateFinalRender                                                                         bool // If the final render position should be recalculated.
	HasInfo                                                                                        bool
	InfoChange                                                                                     bool
	Info                                                                                           []cdata.ObjectInfo
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

// Process is called whenever the object is re-rendered to handle frame advancement and similar.
func (o *Object) Process(dt time.Duration) {
	o.Frame = &(o.Face.Frames[o.FrameIndex])

	// Animate if there are frames and they are visible. NOTE: We *might* want to be able to flag particular animations as requiring having their frames constantly elapsed, or simply record the current real frame and only update the corresponding image render when visibility is restored.
	if len(o.Face.Frames) > 1 && o.Frame.Time > 0 && o.Visible {
		o.FrameElapsed += dt
		for ft := time.Duration(o.Frame.Time) * time.Millisecond; o.FrameElapsed >= ft; {
			o.FrameElapsed -= ft
			o.FrameIndex++
			if o.FrameIndex >= len(o.Face.Frames) {
				o.FrameIndex = 0
			}
			o.Frame = &(o.Face.Frames[o.FrameIndex])
			ft = time.Duration(o.Frame.Time) * time.Millisecond
			o.RecalculateFinalRender = true
		}
	}
}
