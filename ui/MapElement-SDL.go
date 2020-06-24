// +build !mobile

package ui

// MapElement is the element that handles user input and display within a
// field.
type MapElement struct {
	BaseElement
}

// Destroy cleans up the MapElement's resources.
func (m *MapElement) Destroy() {
}

// Render renders the MapElement to the rendering context, with various
// conditionally rendered aspects to represent state.
func (m *MapElement) Render() {
}
