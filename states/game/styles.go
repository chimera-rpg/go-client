package game

var GameContainerStyle string = `
	W 100%
	H 100%
	BackgroundColor 139 186 139 0
`

var MapContainerStyle string = `
	X 20%
	Y 50%
	W 80%
	H 100%
	BackgroundColor 0 0 0 255
	Origin CenterY
`

var DebugWindowStyle string = `
X 20%
Y 0
W 80%
H 100%
BackgroundColor 0 0 0 0
ZIndex 1
`

var ChatWindowStyle string = `
	X 60%
	Y 0
	W 50%
	H 20%
	Origin Bottom CenterX
	BackgroundColor 0 0 0 0
`

var MessagesWindowStyle string = `
	Display Columns
	Direction Reverse
	Origin Bottom
	Y 30
	W 100%
	H 100%
	BackgroundColor 0 0 0 0
`

var CommandContainerStyle string = `
	W 100%
	H 26
	Origin Bottom
	BackgroundColor 0 0 0 32
`

var CommandTypeStyle string = `
	W 10%
	H 100%
`

var ChatInputStyle string = `
	X 10%
	W 85%
`

var InventoryWindowStyle string = `
	X 0
	Y 0
	W 20%
	H 70%
	BackgroundColor 0 128 0 128
`

var InspectorWindowStyle string = `
	X 0
	Y 50%
	W 20%
	H 30%
	BackgroundColor 128 128 0 128
`

var GroundWindowStyle string = `
	X 0
	Y 70%
	W 20%
	H 30%
	BackgroundColor 128 128 128 128
`

var StatsWindowStyle string = `
	X 60%
	Y 0
	W 50%
	H 20%
	Origin CenterX
	BackgroundColor 128 0 0 128
`

var StateWindowStyle string = `
	Origin Right
	X 0
	Y 20%
	W 10%
	H 40%
	BackgroundColor 0 0 0 0
`
