package main

import (
	"fmt"
	"os"

	"github.com/anton2920/imgo"
	"github.com/anton2920/imgo/gr"
)

type AppRect struct {
	imgo.Rect
	Color    imgo.Color
	HotAlpha byte
}

const (
	MaxNumRects = 10

	SelectedNone = -1
)

var (
	Rects        [MaxNumRects]*AppRect
	NumRects     int
	SelectedRect int = SelectedNone
	ActiveRect   *AppRect

	NewRectPos = 120

	DuplicateRectID int
	DeleteRectID    int
	Numeric         bool
	R, G, B         int

	AppUpdateID int
	MakeRectID  int
)

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func GetCreatedRectLocation(window *imgo.Window, size int) int {
	NewRectPos += 20
	if (NewRectPos+size > window.Width()) || (NewRectPos+size > window.Height()) {
		NewRectPos = 100
	}
	return NewRectPos
}

func MakeRect(window *imgo.Window) {
	r := new(AppRect)
	Rects[NumRects] = r

	const size = 100
	location := GetCreatedRectLocation(window, size)
	r.X0 = location
	r.Y0 = location
	r.X1 = r.X0 + size
	r.Y1 = r.Y0 + size

	if SelectedRect != SelectedNone {
		r.Color = Rects[SelectedRect].Color
	} else {
		r.Color = gr.ColorRGB(200, 100, 200)
	}
	r.HotAlpha = 255

	SelectedRect = NumRects
	NumRects++
}

func DuplicateRect() {
	source := Rects[SelectedRect]
	SelectedRect = NumRects
	NumRects++

	r := new(AppRect)
	Rects[SelectedRect] = r

	*r = *source
	r.X0 += 5
	r.Y0 += 5
	r.X1 += 5
	r.Y1 += 5
}

func DeleteRect() {
	if SelectedRect < len(Rects)-1 {
		copy(Rects[SelectedRect:NumRects], Rects[SelectedRect+1:NumRects])
	}
	NumRects--
	SelectedRect = SelectedNone
}

func EditSelection(window *imgo.Window) {
	ui := &window.UI
	r := Rects[SelectedRect]

	if NumRects < MaxNumRects {
		if ui.Button(imgo.ID(&DuplicateRectID), "Duplicate Selected") {
			DuplicateRect()
		}
	}

	ui.ButtonToggle("Show Values", "Hide Values", &Numeric)

	R = int(r.Color.R() >> 4)
	G = int(r.Color.G() >> 4)
	B = int(r.Color.B() >> 4)

	ui.SliderIntDisplay("red", 0, 15, &R, Numeric)
	ui.SliderIntDisplay("green", 0, 15, &G, Numeric)
	ui.SliderIntDisplay("blue", 0, 15, &B, Numeric)

	R = R<<4 + 8
	G = G<<4 + 8
	B = B<<4 + 8

	r.Color = gr.ColorRGB(byte(R), byte(G), byte(B))

	if ui.Button(imgo.ID(&DeleteRectID), "Delete Selected") {
		DeleteRect()
	}
}

func DoRect(window *imgo.Window, n int, fadeRate float32) {
	ui := &window.UI

	handleSize := 5
	r := Rects[n]

	ui.IsHot = false
	ui.IsActive = false

	x := (r.X0 + r.X1) / 2
	y := (r.Y0 + r.Y1) / 2
	xp := x
	yp := y

	if ui.DragXY(imgo.ID2(imgo.ID(r)), &x, abs(r.X1-r.X0), &y, abs(r.Y1-r.Y0)) {
		r.X0 += x - xp
		r.Y0 += y - yp
		r.X1 += x - xp
		r.Y1 += y - yp
	}

	centerHot := ui.IsHot
	ui.IsHot = false

	/* Edges first, so corners are on top. */
	ui.DragX(imgo.ID(&r.X0), &r.X0, handleSize, r.Y0, r.Y1)
	ui.DragX(imgo.ID(&r.X1), &r.X1, handleSize, r.Y0, r.Y1)
	ui.DragY(imgo.ID(&r.Y0), &r.Y0, handleSize, r.X0, r.X1)
	ui.DragY(imgo.ID(&r.Y1), &r.Y1, handleSize, r.X0, r.X1)

	ui.Current.Index = 1 /* change index, so that we can reuse same pointers as new handles .*/
	handleSize = 9       /* corners have larger handles. */

	ui.DragXY(imgo.ID(&r.X0), &r.X0, handleSize, &r.Y0, handleSize)
	ui.DragXY(imgo.ID(&r.Y0), &r.X1, handleSize, &r.Y0, handleSize)
	ui.DragXY(imgo.ID(&r.X1), &r.X0, handleSize, &r.Y1, handleSize)
	ui.DragXY(imgo.ID(&r.Y1), &r.X1, handleSize, &r.Y1, handleSize)

	ui.Current.Index = 0

	if (centerHot) || (ui.IsHot) {
		r.HotAlpha = byte(min(int(float32(r.HotAlpha)+float32(80)*fadeRate), 255))
	} else {
		r.HotAlpha = byte(max(int(float32(r.HotAlpha)-float32(80)*fadeRate), 200))
	}
	ui.Renderer.GraphSolidRect(r.X0, r.Y0, r.X1, r.Y1, (r.Color&0x00FFFFFF)|gr.Color(r.HotAlpha)<<24)

	if ui.IsHot {
		ui.Renderer.GraphRect(r.X0+0, r.Y0+0, r.X1-0, r.Y1-0, gr.ColorBlack)
		ui.Renderer.GraphRect(r.X0+1, r.Y0+1, r.X1-1, r.Y1-1, gr.ColorWhite)
		ui.Renderer.GraphRect(r.X0+2, r.Y0+2, r.X1-2, r.Y1-2, gr.ColorBlack)
	} else {
		ui.Renderer.GraphRect(r.X0+0, r.Y0+0, r.X1-0, r.Y1-0, gr.ColorBlack)
	}

	if ui.IsActive {
		SelectedRect = n
	}
}

func AppUpdate(window *imgo.Window, fadeRate float32) {
	ui := &window.UI
	ui.Begin()

	ui.Renderer.GraphSolidRectWH(0, 0, window.Width(), window.Height(), gr.ColorWhite)

	for i := 0; i < window.Height(); i += 10 {
		ui.Renderer.GraphHLine(i, 0, window.Width(), gr.ColorRGB(225, 255, 230))
	}

	if ui.ButtonLogicRect(imgo.ID(&AppUpdateID), 0, 0, window.Width(), window.Height()) {
		SelectedRect = SelectedNone
	}

	for i := 0; i < NumRects; i++ {
		DoRect(window, i, fadeRate)
	}

	if SelectedRect != SelectedNone {
		if SelectedRect != NumRects-1 {
			r := Rects[SelectedRect]
			copy(Rects[SelectedRect:NumRects], Rects[SelectedRect+1:NumRects])
			SelectedRect = NumRects - 1
			Rects[SelectedRect] = r
		}
	}

	/* Display the panel on the left. Make it grow with the screen width. */
	var width int
	if window.Width() < 400 {
		width = 100
	} else {
		width = window.Width() / 4
	}
	ui.Layout.WidthForAll = width

	if NumRects < MaxNumRects {
		if ui.Button(imgo.ID(&MakeRectID), "Create New") {
			MakeRect(window)
		}
	}

	ui.Layout.CurrentY += 10

	if SelectedRect != SelectedNone {
		EditSelection(window)
	}

	ui.Layout.WidthForAll = imgo.WidthAuto

	/* Done drawing and updating. */
	ui.End()
}

func Fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func init() {
	Rects[0] = new(AppRect)
	Rects[0].X0 = 100
	Rects[0].Y0 = 100
	Rects[0].X1 = 500
	Rects[0].Y1 = 400
	Rects[0].Color = gr.ColorRGB(200, 50, 50)
	Rects[0].HotAlpha = 200

	Rects[1] = new(AppRect)
	Rects[1].X0 = 300
	Rects[1].Y0 = 200
	Rects[1].X1 = 600
	Rects[1].Y1 = 700
	Rects[1].Color = gr.ColorRGB(50, 50, 200)
	Rects[1].HotAlpha = 200

	NumRects = 2
}

func main() {
	window, err := imgo.NewWindow("Immediate Mode GUI demo", 0, 0, 900, 700, imgo.WindowResizable)
	if err != nil {
		Fatal("Failed to create new window: %s\n", err.Error())
	}
	defer window.Close()

	quit := false
	for !quit {
		for window.HasEvents() {
			event := window.GetEvent()
			switch event := event.(type) {
			case imgo.DestroyEvent:
				quit = true
			case imgo.PaintEvent:
				AppUpdate(window, 0)
				window.PaintEvent()
			default:
				window.HandleEvent(event)
			}
		}

		AppUpdate(window, 0.05)
		window.PaintEvent()
	}
}
