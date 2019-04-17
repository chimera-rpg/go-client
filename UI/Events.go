package UI

type Events struct {
	OnTouchBegin      func(id uint32, x int32, y int32) bool
	OnTouchMove       func(id uint32, x int32, y int32) bool
	OnTouchEnd        func(id uint32, x int32, y int32) bool
	OnMouseButtonDown func(id uint32, x int32, y int32) bool
	OnMouseMove       func(id uint32, x int32, y int32) bool
	OnMouseButtonUp   func(id uint32, x int32, y int32) bool
}
