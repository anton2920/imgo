package imgo

import (
	"time"
	"unsafe"

	"github.com/anton2920/imgo/fonts"
	"github.com/anton2920/imgo/gr"
)

type WindowFlags uint

const (
	WindowNone WindowFlags = iota << 1
	WindowHidden
	WindowResizable
	WindowMinimized
	WindowMaximized
)

type Window struct {
	platformWindow
	title  string
	x      int
	y      int
	width  int
	height int
	flags  WindowFlags

	lastPaintEvent time.Time

	UI UI
}

func NewWindow(title string, x, y, width, height int, flags WindowFlags) (*Window, error) {
	w := Window{title: title, x: x, y: y, width: width, height: height, flags: flags}

	if err := platformNewWindow(&w); err != nil {
		return nil, err
	}

	w.UI.Layout = DefaultLayout()
	w.UI.Font = gr.DecompressFont(fonts.Font21)

	w.ResizeEvent(width, height)

	return &w, nil
}

func (w *Window) Close() {
	platformWindowClose(w)
}

func (w *Window) Geometry() Rect {
	return Rect{w.x, w.y, w.x + w.width, w.y + w.height}
}

func (w *Window) GetEvent() interface{} {
	return platformGetEvent(w)
}

func (w *Window) Flags() WindowFlags {
	return w.flags
}

func (w *Window) HasEvents() bool {
	return platformHasEvents(w)
}

func (w *Window) HandleEvent(event Event) {
	switch event := event.(type) {
	case DestroyEvent:
	case MouseButtonDownEvent:
		w.MouseButtonDownEvent(event.X, event.Y, event.Button)
	case MouseButtonUpEvent:
		w.MouseButtonUpEvent(event.X, event.Y, event.Button)
	case MouseMoveEvent:
		w.MouseMoveEvent(event.X, event.Y)
	case PaintEvent:
		w.PaintEvent()
	case ResizeEvent:
		w.ResizeEvent(event.Width, event.Height)
	}
}

func (w *Window) Height() int {
	return w.height
}

func (w *Window) Invalidate() {
	platformRedrawAll(w)
}

func (w *Window) MouseButtonDownEvent(x, y int, button MouseButton) {
	w.UI.MouseX = x
	w.UI.MouseY = y

	switch button {
	case Button1:
		w.UI.LeftDown = true
		/* TODO(anton2920): is this function really necessary? */
		platformMouseCapture(w, true)
	case Button2:
		w.UI.MiddleDown = true
	case Button3:
		w.UI.RightDown = true
	}
}

func (w *Window) MouseButtonUpEvent(x, y int, button MouseButton) {
	w.UI.MouseX = x
	w.UI.MouseY = y

	switch button {
	case Button1:
		w.UI.LeftUp = true
		/* TODO(anton2920): is this function really necessary? */
		platformMouseCapture(w, false)
	case Button2:
		w.UI.MiddleUp = true
	case Button3:
		w.UI.RightUp = true
	}
}

func (w *Window) MouseMoveEvent(x, y int) {
	w.UI.MouseX = x
	w.UI.MouseY = y
}

func (w *Window) PaintEvent() {
	screen := w.UI.Renderer.pixmap
	platformDrawPixmap(w, 0, 0, unsafe.Slice((*RGBA)(unsafe.Pointer(unsafe.SliceData(screen.Pixels))), screen.Width*screen.Height), screen.Width, screen.Height, screen.Stride)

	/* Preventing CPU from going wild. */
	const FPS = 60
	now := time.Now()
	durationBetweenPaints := now.Sub(w.lastPaintEvent)
	if durationBetweenPaints < 1000/FPS*time.Millisecond {
		//time.Sleep(1000/FPS*time.Millisecond - durationBetweenPaints)
	}
	w.lastPaintEvent = time.Now()
}

func (w *Window) ResizeEvent(width, height int) {
	w.width = width
	w.height = height

	screen := w.UI.Renderer.pixmap
	if (width > screen.Width) || (height > screen.Height) {
		screen = gr.NewPixmap(width, height, gr.AlphaOpaque)
	} else {
		screen.Pixels = screen.Pixels[:width*height]
		screen.Width = width
		screen.Height = height
		screen.Stride = width
	}
	w.UI.Renderer.pixmap = screen
	w.UI.Renderer.active = Rect{0, 0, screen.Width, screen.Height}

	w.Invalidate()
}

func (w *Window) Title() string {
	return w.title
}

func (w *Window) Width() int {
	return w.width
}

func (w *Window) X() int {
	return w.x
}

func (w *Window) Y() int {
	return w.y
}
