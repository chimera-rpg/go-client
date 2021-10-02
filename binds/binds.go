package binds

// Bindings represent a structure for managing and triggering binds.
type Bindings struct {
	Keygroups map[string][]KeyGroup
	functions map[string]func(i ...interface{})
}

// NewBindings returns a constructed Bindings.
func NewBindings() *Bindings {
	return &Bindings{
		Keygroups: make(map[string][]KeyGroup),
		functions: make(map[string]func(...interface{})),
	}
}

// Init ensures the Bindings structure is initialized.
func (b *Bindings) Init() {
	if b.Keygroups == nil {
		b.Keygroups = make(map[string][]KeyGroup)
	}
	if b.functions == nil {
		b.functions = make(map[string]func(...interface{}))
	}
}

// Trigger calls any bound functions that are tied to the given keygroup.
func (b *Bindings) Trigger(k KeyGroup, i ...interface{}) {
	var triggers []string
	for name, keygroups := range b.Keygroups {
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

// RunFunction attempts to run the function associated with name
func (b *Bindings) RunFunction(name string, i ...interface{}) {
	if f, ok := b.functions[name]; ok {
		f(i...)
	}
}

// SetFunction sets the associated bind name to a function.
func (b *Bindings) SetFunction(name string, f func(i ...interface{})) {
	b.functions[name] = f
}

// HasFunction returns if there is a bound function matching name.
func (b *Bindings) HasFunction(name string) bool {
	_, ok := b.functions[name]
	return ok
}

// HasKeygroupsForName returns if there are any keygroups matching the given name.
func (b *Bindings) HasKeygroupsForName(name string) bool {
	_, ok := b.Keygroups[name]
	return ok
}

// AddKeygroup adds a keygroup.
func (b *Bindings) AddKeygroup(name string, k KeyGroup) {
	if b.FindKeyGroupIndex(name, k) == -1 {
		b.Keygroups[name] = append(b.Keygroups[name], k)
	}
}

// RemoveKeygroup removes a keygroup.
func (b *Bindings) RemoveKeygroup(name string, k KeyGroup) {
	i := b.FindKeyGroupIndex(name, k)
	if i != -1 {
		b.Keygroups[name] = append(b.Keygroups[name][:i], b.Keygroups[name][i+1:]...)
	}
}

// FindKeyGroupIndex finds a key group entry matching the given one, returning its position. -1 indicates missing.
func (b *Bindings) FindKeyGroupIndex(name string, k KeyGroup) int {
	if arr, ok := b.Keygroups[name]; ok {
		for i, v := range arr {
			if k.Same(v) {
				return i
			}
		}
	}
	return -1
}
