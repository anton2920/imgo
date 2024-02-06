package main

import (
	"fmt"
	"os"

	"github.com/anton2920/imgo"
	"github.com/anton2920/imgo/gr"
)

func GridExample(window *imgo.Window) {
	ui := &window.UI

	ui.Begin()

	ui.Renderer.GraphSolidRectWH(0, 0, window.Width(), window.Height(), gr.ColorGrey(220))

	const rectsPerRow = 3
	const rectsPerCol = 3
	const MarginLeft = 10
	const MarginTop = 10
	const paddingLeft = 5
	const paddingTop = 5
	rectWidth := (window.Width() - MarginLeft*2) / rectsPerRow
	rectHeight := (window.Height() - MarginTop*2) / rectsPerCol

	for i := 0; i < rectsPerRow; i++ {
		for j := 0; j < rectsPerCol; j++ {
			x := (MarginLeft + i*rectWidth) + paddingLeft
			y := (MarginTop + j*rectHeight) + paddingTop
			width := rectWidth - paddingLeft
			height := rectHeight - paddingTop
			color := gr.ColorRGB(byte(float32(i)*255/rectsPerRow), byte(float32(i)*float32(j)*255/(rectsPerRow*rectsPerCol)), byte(float32(j)*255/rectsPerCol))

			ui.IsHot = false
			ui.IsActive = false

			if ui.ButtonLogicRect(imgo.ID(uintptr(i*rectsPerRow+j)+1), x, y, width, height) {
				println(x, y, width, height, "PRESS!", ui.MouseX, ui.MouseY)
			}
			if !ui.IsHot {
				color = gr.ColorRGBA(color.R(), color.G(), color.B(), 150)
			}
			ui.Renderer.GraphSolidRectWH(x, y, width, height, color)
		}
	}

	ui.End()
}

func main() {
	window, err := imgo.NewWindow("Grid", 0, 0, 600, 600, imgo.WindowResizable)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create new window: %s\n", err.Error())
		os.Exit(1)
	}
	defer window.Close()

	quit := false
	for !quit {
		for window.HasEvents() {
			event := window.GetEvent()
			switch event := event.(type) {
			case imgo.DestroyEvent:
				quit = true
			default:
				window.HandleEvent(event)
			}

			GridExample(window)
			window.PaintEvent()
		}
	}
}
