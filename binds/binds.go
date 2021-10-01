package binds

// Bindings represent a structure for managing and triggering binds.
type Bindings struct {
	keygroups map[string][]KeyGroup
	functions map[string]func(i ...interface{})
}

// NewBindings returns a constructed Bindings.
func NewBindings() *Bindings {
	return &Bindings{
		keygroups: make(map[string][]KeyGroup),
		functions: make(map[string]func(...interface{})),
	}
}

// Trigger calls any bound functions that are tied to the given keygroup.
func (b *Bindings) Trigger(k KeyGroup, i ...interface{}) {
	var triggers []string
	for name, keygroups := range b.keygroups {
		for _, keygroup := range keygroups {
			if keygroup.Same(k) {
				triggers = append(triggers, name)
			}
		}
	}
	for _, trigger := range triggers {
		if f, ok := b.functions[trigger]; ok {
			f(i...)
		}
	}
}

// SetFunction sets the associated bind name to a function.
func (b *Bindings) SetFunction(name string, f func(i ...interface{})) {
	b.functions[name] = f
}

// AddKeygroup adds a keygroup.
func (b *Bindings) AddKeygroup(name string, k KeyGroup) {
	if b.FindKeyGroupIndex(name, k) == -1 {
		b.keygroups[name] = append(b.keygroups[name], k)
	}
}

// RemoveKeygroup removes a keygroup.
func (b *Bindings) RemoveKeygroup(name string, k KeyGroup) {
	i := b.FindKeyGroupIndex(name, k)
	if i != -1 {
		b.keygroups[name] = append(b.keygroups[name][:i], b.keygroups[name][i+1:]...)
	}
}

// FindKeyGroupIndex finds a key group entry matching the given one, returning its position. -1 indicates missing.
func (b *Bindings) FindKeyGroupIndex(name string, k KeyGroup) int {
	if arr, ok := b.keygroups[name]; ok {
		for i, v := range arr {
			if k.Same(v) {
				return i
			}
		}
	}
	return -1
}
