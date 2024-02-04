package imgo

type MouseButton int

const (
	Button1 MouseButton = iota + 1
	Button2
	Button3
	Button4
	Button5
)

type Event interface{}

type DestroyEvent struct{}

type MouseButtonDownEvent struct {
	X, Y   int
	Button MouseButton
}

type MouseButtonUpEvent struct {
	X, Y   int
	Button MouseButton
}

type MouseMoveEvent struct {
	X, Y int
}

type PaintEvent struct{}

type ResizeEvent struct {
	Width, Height int
}
