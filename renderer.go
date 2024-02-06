package imgo

import "github.com/anton2920/imgo/gr"

type surface struct {
	pixmap gr.Pixmap
	active gr.Rect
}

type Renderer surface

func (r *Renderer) GraphCircle(x0, y0, radius int, color Color) {
	gr.DrawCircle(r.pixmap, r.active, x0, y0, radius, color)
}

func (r *Renderer) GraphHLine(y, x0, x1 int, color Color) {
	gr.DrawHLine(r.pixmap, r.active, y, x0, x1, color)
}

func (r *Renderer) GraphRect(x0, y0, x1, y1 int, color Color) {
	gr.DrawRectOutline(r.pixmap, r.active, x0, y0, x1, y1, color)
}

func (r *Renderer) GraphRectWH(x, y, width, height int, color Color) {
	gr.DrawRectOutlineWH(r.pixmap, r.active, x, y, width, height, color)
}

func (r *Renderer) GraphPoint(x, y, size int, color Color) {
	if size <= 1 {
		gr.DrawPoint(r.pixmap, r.active, x, y, color)
	} else {
		gr.DrawRectSolid(r.pixmap, r.active, x-size, y-size, x+size, y+size, color)
	}
}

func (r *Renderer) GraphSolidRect(x0, y0, x1, y1 int, color Color) {
	gr.DrawRectSolid(r.pixmap, r.active, x0, y0, x1, y1, color)
}

func (r *Renderer) GraphSolidRectWH(x, y, width, height int, color Color) {
	gr.DrawRectSolidWH(r.pixmap, r.active, x, y, width, height, color)
}

func (r *Renderer) GraphText(text string, font Font, x, y int, color Color) {
	gr.DrawText(r.pixmap, r.active, text, font, x, y, color)
}
