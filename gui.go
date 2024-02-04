package imgo

type GUI struct {
	GR Graphics

	regionStack [128]struct {
		previous Pixmap
		x, y     int
		ox, oy   int
	}
	stackIndex int
	dx, dy     int
}

func (gui *GUI) GraphRect(x0, y0, x1, y1 int, color Color) {
	dx := gui.dx
	dy := gui.dy
	gui.GR.DrawRectOutline(x0-dx, y0-dy, x1-dx, y1-dy, color)
}

func (gui *GUI) GraphRectWH(x, y, width, height int, color Color) {
	dx := gui.dx
	dy := gui.dy
	gui.GR.DrawRectOutlineWH(x-dx, y-dy, width, height, color)
}

func (gui *GUI) GraphPoint(x, y, size int, color Color) {
	x += gui.dx
	y += gui.dy
	if size <= 1 {
		gui.GR.DrawPoint(x, y, color)
	} else {
		gui.GR.DrawRectSolid(x-size, y-size, x+size, y+size, color)
	}
}

func (gui *GUI) GraphSolidRect(x0, y0, x1, y1 int, color Color) {
	dx := gui.dx
	dy := gui.dy
	gui.GR.DrawRectSolid(x0-dx, y0-dy, x1-dx, y1-dy, color)
}

func (gui *GUI) GraphSolidRectWH(x, y, width, height int, color Color) {
	dx := gui.dx
	dy := gui.dy
	gui.GR.DrawRectSolidWH(x-dx, y-dy, width, height, color)
}

func (gui *GUI) GraphText(text string, x, y int, color Color) {
	dx := gui.dx
	dy := gui.dy
	gui.GR.Text(text, x-dx, y-dy, color)
}
