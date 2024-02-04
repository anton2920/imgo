package main

import (
	"fmt"
	"os"

	"github.com/anton2920/imgo"
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
	ui *imgo.UI

	Rects        [MaxNumRects]*AppRect
	NumRects     int
	SelectedRect int = SelectedNone

	ActiveRect *AppRect

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

func GetCreatedRectLocation(size int) int {
	screen := ui.GUI.GR.GetOutput()

	NewRectPos += 20
	if (NewRectPos+size > screen.Width) || (NewRectPos+size > screen.Height) {
		NewRectPos = 100
	}

	return NewRectPos
}

func MakeRect() {
	r := new(AppRect)
	Rects[NumRects] = r

	const size = 100
	location := GetCreatedRectLocation(size)
	r.X0 = location
	r.Y0 = location
	r.X1 = r.X0 + size
	r.Y1 = r.Y0 + size

	if SelectedRect != SelectedNone {
		r.Color = Rects[SelectedRect].Color
	} else {
		r.Color = imgo.ColorRGB(200, 100, 200)
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

func EditSelection() {
	r := Rects[SelectedRect]

	if NumRects < MaxNumRects {
		if ui.Button(imgo.ID(&DuplicateRectID), "Duplicate Selected") {
			DuplicateRect()
		}
	}

	ui.ButtonToggle("Show Values", "Hide Values", &Numeric)

	R = int(r.Color.RedComponent() >> 4)
	G = int(r.Color.GreenComponent() >> 4)
	B = int(r.Color.BlueComponent() >> 4)

	ui.SliderIntDisplay("red", 0, 15, &R, Numeric)
	ui.SliderIntDisplay("green", 0, 15, &G, Numeric)
	ui.SliderIntDisplay("blue", 0, 15, &B, Numeric)

	R = R<<4 + 8
	G = G<<4 + 8
	B = B<<4 + 8

	r.Color = imgo.ColorRGB(byte(R), byte(G), byte(B))

	if ui.Button(imgo.ID(&DeleteRectID), "Delete Selected") {
		DeleteRect()
	}
}

func DoRect(n int, fadeRate float32) {
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
	ui.GUI.GraphSolidRect(r.X0, r.Y0, r.X1, r.Y1, (r.Color&0x00FFFFFF)|imgo.Color(r.HotAlpha)<<24)

	if ui.IsHot {
		ui.GUI.GraphRect(r.X0+0, r.Y0+0, r.X1-0, r.Y1-0, imgo.ColorBlack)
		ui.GUI.GraphRect(r.X0+1, r.Y0+1, r.X1-1, r.Y1-1, imgo.ColorWhite)
		ui.GUI.GraphRect(r.X0+2, r.Y0+2, r.X1-2, r.Y1-2, imgo.ColorBlack)
	} else {
		ui.GUI.GraphRect(r.X0+0, r.Y0+0, r.X1-0, r.Y1-0, imgo.ColorBlack)
	}

	if ui.IsActive {
		SelectedRect = n
	}
}

func AppUpdate(fadeRate float32) {
	ui.Begin()

	screen := ui.GUI.GR.GetOutput()
	ui.GUI.GR.DrawRectSolid(0, 0, screen.Width, screen.Height, imgo.ColorWhite)

	for i := 0; i < screen.Height; i += 10 {
		ui.GUI.GR.DrawHLine(i, 0, screen.Width, imgo.ColorRGB(225, 255, 230))
	}

	if ui.ButtonLogicRect(imgo.ID(&AppUpdateID), 0, 0, screen.Width, screen.Height) {
		SelectedRect = SelectedNone
	}

	for i := 0; i < NumRects; i++ {
		DoRect(i, fadeRate)
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
	if screen.Width < 400 {
		width = 100
	} else {
		width = screen.Width / 4
	}
	ui.Layout.WidthForAll = width

	if NumRects < MaxNumRects {
		if ui.Button(imgo.ID(&MakeRectID), "Create New") {
			MakeRect()
		}
	}

	ui.Layout.CurrentY += 10

	if SelectedRect != SelectedNone {
		EditSelection()
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
	Rects[0].Color = imgo.ColorRGB(200, 50, 50)
	Rects[0].HotAlpha = 200

	Rects[1] = new(AppRect)
	Rects[1].X0 = 300
	Rects[1].Y0 = 200
	Rects[1].X1 = 600
	Rects[1].Y1 = 700
	Rects[1].Color = imgo.ColorRGB(50, 50, 200)
	Rects[1].HotAlpha = 200

	NumRects = 2
}

func main() {
	window, err := imgo.NewWindow("Immediate Mode GUI demo", 0, 0, 900, 700, imgo.WindowResizable)
	if err != nil {
		Fatal("Failed to create new window: %s\n", err.Error())
	}
	defer window.Close()

	ui = &window.UI

	quit := false
	for !quit {
		for window.HasEvents() {
			event := window.GetEvent()
			switch event := event.(type) {
			case imgo.DestroyEvent:
				quit = true
			case imgo.PaintEvent:
				AppUpdate(0)
				window.PaintEvent()
			default:
				window.HandleEvent(event)
			}
		}

		AppUpdate(0.05)
		window.PaintEvent()
	}
}
