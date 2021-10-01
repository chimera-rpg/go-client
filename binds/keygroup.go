package binds

// KeyGroup provides a container for modifiers + keys
type KeyGroup struct {
	Keys      []uint8
	Modifiers uint16
	Pressed   bool
	Repeat    bool
}

// Same returns whether or not the keygroup is the same as another.
func (k *KeyGroup) Same(o KeyGroup) bool {
	if len(k.Keys) != len(o.Keys) {
		return false
	}
	if k.Modifiers != o.Modifiers {
		return false
	}
	if k.Pressed != o.Pressed {
		return false
	}
	if k.Repeat != o.Repeat {
		return false
	}
	for i := 0; i < len(k.Keys); i++ {
		if k.Keys[i] != o.Keys[i] {
			return false
		}
	}
	return true
}
