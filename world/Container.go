package world

// Container represents a single container. 0 is always the player's base inventory.
type Container struct {
	ID        uint32
	ObjectIDs []uint32
}

// UpdateContainer updates the given container with the list of objects.
func (w *World) UpdateContainer(ID uint32, objects []uint32) {
	var container *Container
	for _, c := range w.containers {
		if c.ID == ID {
			container = c
			break
		}
	}
	if container == nil {
		container = &Container{
			ID: ID,
		}
		w.containers = append(w.containers, container)
	}

	// Automatically set any objects no longer in the container as not contained.
	for _, oID := range container.ObjectIDs {
		has := false
		for _, oID2 := range objects {
			if oID2 == oID {
				has = true
				break
			}
		}
		if !has {
			if o := w.GetObject(oID); o != nil {
				o.Contained = false
			}
		}
	}

	container.ObjectIDs = objects
	// Create objects as necessary.
	for _, oID := range container.ObjectIDs {
		o := w.GetObject(oID)
		if o == nil {
			w.AddObject(&Object{
				ID: oID,
			})
			o = w.GetObject(oID)
		}
		o.Contained = true
	}
}
