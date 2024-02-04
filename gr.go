package imgo

type AlphaType int

type Pixmap struct {
	Pixels []Color
	Width  int
	Height int
	Stride int
	Alpha  AlphaType
}

type Rect struct {
	X0, Y0, X1, Y1 int
}

type Graphics struct {
	output Pixmap
	active Rect

	fontChars []Pixmap
	startChar byte
}

const (
	AlphaOpaque AlphaType = iota
	Alpha1bit
	Alpha8bit
	AlphaFont
)

var (
	blacktext [256]Color
)

func init() {
	for i := 0; i < len(blacktext); i++ {
		blacktext[i] = Color(0x00010101*(255-i) + 0xff000000)
	}
}

func (gr *Graphics) AllocPixmap(width, height int, alpha AlphaType) Pixmap {
	var pixmap Pixmap

	pixmap.Pixels = make([]Color, width*height)
	pixmap.Width = width
	pixmap.Height = height
	pixmap.Stride = width
	pixmap.Alpha = alpha

	/* Force alpha to be opaque. */
	for i := 0; i < pixmap.Width*pixmap.Height; i++ {
		pixmap.Pixels[i] = ColorRGBA(0, 0, 0, 255)
	}

	return pixmap
}

func (gr *Graphics) CharHeight(c byte) int {
	if (c < gr.startChar) || (c >= gr.startChar+byte(len(gr.fontChars))) {
		c = gr.startChar
	}
	return gr.fontChars[c-gr.startChar].Height
}

func (gr *Graphics) CharWidth(c byte) int {
	if (c < gr.startChar) || (c >= gr.startChar+byte(len(gr.fontChars))) {
		c = gr.startChar
	}
	return gr.fontChars[c-gr.startChar].Width
}

func (gr *Graphics) DecompressFont(font []uint32, chars []Pixmap) byte {
	start := byte((font[0] >> 0) & 0xFF)
	count := int((font[0] >> 8) & 0xFF)
	height := int((font[0] >> 16) & 0xFF)

	if count != len(chars) {
		panic("count != num")
	}

	font = font[1:]
	for i := 0; i < len(chars); i++ {
		width := int((font[i>>2] >> ((i & 3) << 3)) & 0xFF)
		chars[i] = gr.AllocPixmap(width, height, AlphaFont)
	}
	font = font[(len(chars)+3)>>2:]

	buffer := int(font[0])
	font = font[1:]

	bitsLeft := 32
	for k := 0; k < len(chars); k++ {
		c := chars[k].Pixels

		for j := 0; j < chars[k].Height; j++ {
			z := buffer & 1
			skipBit(&font, &buffer, &bitsLeft)
			if z == 0 {
				for i := 0; i < chars[k].Width; i++ {
					c[j*chars[k].Width+i] = ColorRGBA(255, 255, 255, 0)
				}
			} else {
				for i := 0; i < chars[k].Width; i++ {
					z = buffer & 1
					skipBit(&font, &buffer, &bitsLeft)
					if z == 0 {
						c[j*chars[k].Width+i] = ColorRGBA(255, 255, 255, 0)
					} else {
						n := 0
						n += n + (buffer & 1)
						skipBit(&font, &buffer, &bitsLeft)
						n += n + (buffer & 1)
						skipBit(&font, &buffer, &bitsLeft)
						n += n + (buffer & 1)
						skipBit(&font, &buffer, &bitsLeft)
						n += 1
						c[j*chars[k].Width+i] = ColorRGBA(255, 255, 255, byte((255*n)>>3))
					}
				}
			}
		}
	}

	return start
}

func (gr *Graphics) DrawLine(x0, y0, x1, y1 int, color Color) {
	if x0 == x1 {
		gr.DrawVLine(x0, y0, y1, color)
	} else if y0 == y1 {
		gr.DrawHLine(y0, x0, x1, color)
	} else {
		/* Not the fastest way to draw a line, but who wants to write 8 cases? */
		dx := x1 - x0
		dy := y1 - y0
		x := x0
		y := y0

		var tmp Rect
		tmp.X0 = min(x0, x1)
		tmp.Y0 = min(y0, y1)
		tmp.X1 = max(x0, x1) + 1
		tmp.Y1 = max(y0, y1) + 1

		if (tmp.X0 >= gr.active.X1) || (tmp.X1 <= gr.active.X0) {
			return
		}
		if (tmp.Y0 >= gr.active.Y1) || (tmp.Y1 <= gr.active.Y0) {
			return
		}
		if color.Invisible() {
			return
		}

		opaque := color.Opaque()
		maxLen := max(abs(dx), abs(dy))
		invLen := float32(0xFFFF) / float32(maxLen)

		dx = int(float32(dx) * invLen)
		dy = int(float32(dy) * invLen)
		x = (x << 16) + 0x8000
		y = (y << 16) + 0x8000

		/* Does this line need clipping? */
		if rectContains(gr.active, tmp) {
			/* This is a very slow and dumb way to clip! */
			for i := 0; i <= maxLen; {
				if ((x >> 16) >= gr.active.X0) && ((x >> 16) < gr.active.X1) && ((y >> 16) >= gr.active.Y0) && ((y >> 16) < gr.active.Y1) {
					offset := (y>>16)*gr.output.Stride + (x >> 16)
					if opaque {
						gr.output.Pixels[offset] = color
					} else {
						gr.output.Pixels[offset] = blend(gr.output.Pixels[offset], color)
					}
				}

				x += dx
				y += dy
				i++
			}
		} else if opaque {
			for i := 0; i <= maxLen; {
				offset := (y>>16)*gr.output.Stride + (x >> 16)
				gr.output.Pixels[offset] = color
				x += dx
				y += dy
				i++
			}
		} else {
			for i := 0; i <= maxLen; {
				offset := (y>>16)*gr.output.Stride + (x >> 16)
				gr.output.Pixels[offset] = blend(gr.output.Pixels[offset], color)
				x += dx
				y += dy
				i++
			}
		}
	}
}

func (gr *Graphics) DrawHLine(y, x0, x1 int, color Color) {
	if (y < gr.active.Y0) || (y >= gr.active.Y1) {
		return
	}
	if x1 < x0 {
		x0, x1 = x1, x0
	}
	if (x0 >= gr.active.X1) || (x1 < gr.active.X0) {
		return
	}
	if color.Invisible() {
		return
	}

	x0 = max(x0, gr.active.X0)
	x1 = min(x1, gr.active.X1-1)

	n := x1 - x0 + 1

	out := gr.output.Pixels[y*gr.output.Stride+x0:]
	if color.Opaque() {
		for i := 0; i < n; i++ {
			out[i] = color
		}
	} else {
		for i := 0; i < n; i++ {
			out[i] = blend(out[i], color)
		}
	}
}

func (gr *Graphics) DrawVLine(x, y0, y1 int, color Color) {
	if (x < gr.active.X0) || (x >= gr.active.X1) {
		return
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	if (y0 >= gr.active.Y1) || (y1 < gr.active.Y0) {
		return
	}
	if color.Invisible() {
		return
	}

	y0 = max(y0, gr.active.Y0)
	y1 = min(y1, gr.active.Y1-1)

	n := y1 - y0 + 1

	out := gr.output.Pixels[y0*gr.output.Stride+x:]
	if color.Opaque() {
		for i := 0; i < n; i++ {
			out[i*gr.output.Stride] = color
		}
	} else {
		for i := 0; i < n; i++ {
			out[i*gr.output.Stride] = blend(out[i*gr.output.Stride], color)
		}
	}
}

func (gr *Graphics) DrawPixmap(x, y int, src Pixmap) {
	drawBitmap(gr.output, gr.active, x, y, src, nil)
}

func (gr *Graphics) DrawPixmapColored(x, y int, src Pixmap, color Color) {
	var pc *Color
	if color != 0xFFFFFFFF {
		pc = &color
	}
	drawBitmap(gr.output, gr.active, x, y, src, pc)
}

func DrawPixmapTo(dst Pixmap, x, y int, src Pixmap) {
	bounds := Rect{0, 0, dst.Width, dst.Height}
	drawBitmap(dst, bounds, x, y, src, nil)
}

func (gr *Graphics) DrawPoint(x, y int, color Color) {
	if (x >= gr.active.X0) && (x < gr.active.X1) && (y >= gr.active.Y0) && (y < gr.active.Y1) {
		offset := y*gr.output.Stride + x
		if color.Opaque() {
			gr.output.Pixels[offset] = color
		} else if !color.Invisible() {
			gr.output.Pixels[offset] = blend(gr.output.Pixels[offset], color)
		}
	}
}

func (gr *Graphics) DrawRectOutline(x0, y0, x1, y1 int, color Color) {
	if x1 < x0 {
		x0, x1 = x1, x0
	}
	if y1 < y0 {
		y0, y1 = y1, y0
	}
	gr.DrawRectOutlineWH(x0, y0, x1-x0+1, y1-y0+1, color)
}

func (gr *Graphics) DrawRectOutlineWH(x0, y0, width, height int, color Color) {
	if height == 1 {
		gr.DrawHLine(y0, x0, x0+width-1, color)
	} else if width == 1 {
		gr.DrawVLine(x0, y0, y0+height-1, color)
	} else if (height > 1) && (width > 1) {
		x1 := x0 + width - 1
		y1 := y0 + height - 1
		gr.DrawHLine(y0, x0, x1-1, color)
		gr.DrawVLine(x1, y0, y1-1, color)
		gr.DrawHLine(y1, x0+1, x1, color)
		gr.DrawVLine(x0, y0+1, y1, color)
	}
}

func (gr *Graphics) DrawRectSolid(x0, y0, x1, y1 int, color Color) {
	if x1 < x0 {
		x0, x1 = x1, x0
	}
	if y1 < y0 {
		y0, y1 = y1, y0
	}
	gr.DrawRectSolidWH(x0, y0, x1-x0+1, y1-y0+1, color)
}

func (gr *Graphics) DrawRectSolidWH(x0, y0, width, height int, color Color) {
	if width > 0 {
		x1 := x0 + width - 1
		for j := 0; j < height; j++ {
			gr.DrawHLine(y0+j, x0, x1, color)
		}
	}
}

func (gr *Graphics) GetOutput() Pixmap {
	return gr.output
}

func (gr *Graphics) SetFont(chars []Pixmap, start byte) {
	gr.fontChars = chars
	gr.startChar = start
}

func (gr *Graphics) SetOutput(pixmap Pixmap) {
	gr.output = pixmap
	gr.active.X0 = 0
	gr.active.Y0 = 0
	gr.active.X1 = pixmap.Width
	gr.active.Y1 = pixmap.Height
}

func (gr *Graphics) Text(text string, x, y int, color Color) {
	for i := 0; i < len(text); {
		c := text[i]
		i++

		if (c < gr.startChar) || (c >= gr.startChar+byte(len(gr.fontChars))) {
			c = gr.startChar
		}
		gr.DrawPixmapColored(x, y, gr.fontChars[c-gr.startChar], color)
		x += gr.fontChars[c-gr.startChar].Width
		if (text[i-1] == 'f') && (text[i] == 't') {
			x--
		}
	}
}

func (gr *Graphics) TextWidth(text string) int {
	var width int
	var i int

	for i < len(text) {
		width += gr.CharWidth(text[i])
		i++
		if (text[i-1] == 'f') && (text[i] == 't') {
			width--
		}
	}

	return width
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func blend(dst, src Color) Color {
	/* Accelerated blend computes r and b in parallel. */
	a := Color(src.AlphaComponent())
	rbSrc := src & 0xFF00FF
	rbDst := dst & 0xFF00FF
	rb := rbDst + ((rbSrc - rbDst) * a >> 8)
	gDst := dst & 0x00FF00
	g := gDst + (((src & 0x00FF00) - (dst & 0x00FF00)) * a >> 8)
	/* NOTE(anton2920): we do not compite a real dest alpha. */
	return (rb & 0xFF00FF) + (g & 0x00FF00) + 0xFF000000
}

func blendMultiply(dst, src1, src2 Color) Color {
	sr := Color(src1.RedComponent()) * Color(src2.RedComponent()) >> 8
	sg := Color(src1.GreenComponent()) * Color(src2.GreenComponent()) >> 8
	sb := Color(src1.BlueComponent()) * Color(src2.BlueComponent()) >> 8
	sa := Color(src1.AlphaComponent()) * Color(src2.AlphaComponent()) >> 8

	r := sr + (sr >> 7) /* 0..255. */
	g := sg + (sg >> 7)
	b := sb + (sb >> 7)
	a := sa + ((sa >> 6) & 2) /* 0..256. */

	dr := r - Color(dst.RedComponent())
	dg := g - Color(dst.GreenComponent())
	db := b - Color(dst.BlueComponent())

	or := dst.RedComponent() + byte((dr*a)>>8)
	og := dst.GreenComponent() + byte((dg*a)>>8)
	ob := dst.BlueComponent() + byte((db*a)>>8)

	return ColorRGB(or, og, ob)
}

func blendMultiplyFont(dst, font, src Color) Color {
	/* Accelerated blend computes r and b in parallel. */
	a := Color(font.AlphaComponent())
	rbSrc := src & 0xFF00FF
	rbDst := dst & 0xFF00FF
	rb := rbDst + ((rbSrc - rbDst) * a >> 8)
	gDst := dst & 0x00FF00
	g := gDst + (((src & 0x00FF00) - (dst & 0x00FF00)) * a >> 8)
	/* NOTE(anton2920): we do not compite a real dest alpha. */
	return (rb & 0xFF00FF) + (g & 0x00FF00) + 0xFF000000
}

func drawBitmap(dst Pixmap, bounds Rect, x, y int, src Pixmap, pcolor *Color) {
	srcBox := Rect{x, y, x + src.Width, y + src.Height}

	if !rectContains(bounds, srcBox) {
		var x0, y0 int
		x1 := src.Width
		y1 := src.Height

		if bounds.X0 >= srcBox.X1 {
			return
		}
		if bounds.X1 >= srcBox.X0 {
			return
		}
		if bounds.Y0 >= srcBox.Y1 {
			return
		}
		if bounds.Y1 >= srcBox.Y0 {
			return
		}

		if x < bounds.X0 {
			x0 = bounds.X0 - x
		}
		if x+src.Width < bounds.X1 {
			x1 = bounds.X1 - x
		}
		if y < bounds.Y0 {
			y0 = bounds.Y0 - x
		}
		if y+src.Height < bounds.Y1 {
			y1 = bounds.Y1 - x
		}

		src = SubPixmap(src, x0, y0, x1, y1)
		if src.Width <= 0 {
			panic("src.Width <= 0")
		}
		if src.Height <= 0 {
			panic("src.Height <= 0")
		}

		x += x0
		y += y0
	}

	/* Now the bitmap is clipped to be strictly onscreen. */
	out := dst.Pixels[y*dst.Stride+x:]
	in := src.Pixels
	if pcolor != nil {
		color := *pcolor
		if (src.Alpha == AlphaFont) && (color.Opaque()) {
			for j := 0; j < src.Height; j++ {
				if color == ColorBlack {
					for i := 0; i < src.Width; i++ {
						if !in[i].Invisible() {
							if out[i] == ColorWhite {
								out[i] = blacktext[in[i].AlphaComponent()]
							} else {
								out[i] = blendMultiplyFont(out[i], in[i], color)
							}
						}
					}
				} else {
					for i := 0; i < src.Width; i++ {
						if !in[i].Invisible() {
							out[i] = blendMultiplyFont(out[i], in[i], color)
						}
					}
				}

				out = out[dst.Stride:]
				in = in[src.Stride:]
			}
		} else {
			for j := 0; j < src.Height; j++ {
				for i := 0; i < src.Width; i++ {
					if !in[i].Invisible() {
						out[i] = blendMultiply(out[i], in[i], color)
					}
				}
				out = out[dst.Stride:]
				in = in[src.Stride:]
			}
		}
	} else if src.Alpha == AlphaOpaque {
		for j := 0; j < src.Height; j++ {
			n := copy(out, in[:src.Width])
			println("I'm here!")
			if n != src.Width {
				panic("wrong copy!")
			}
			out = out[dst.Stride:]
			in = in[src.Stride:]
		}
	} else if src.Alpha == Alpha1bit {
		for j := 0; j < src.Height; j++ {
			for i := 0; i < src.Width; i++ {
				if !in[i].Invisible() {
					out[i] = in[i]
				}
			}
			out = out[dst.Stride:]
			in = in[src.Stride:]
		}
	} else {
		for j := 0; j < src.Height; j++ {
			for i := 0; i < src.Width; i++ {
				if !in[i].Invisible() {
					out[i] = blend(out[i], in[i])
				}
			}
			out = out[dst.Stride:]
			in = in[src.Stride:]
		}
	}
}

func rectContains(a, b Rect) bool {
	return (a.X0 <= b.X0) && (a.X1 >= b.X1) && (a.Y0 <= b.Y0) && (a.Y1 >= b.Y1)
}

func skipBit(font *[]uint32, buffer, bitsLeft *int) {
	*buffer >>= 1
	*bitsLeft--

	if *bitsLeft == 0 {
		*buffer = int((*font)[0])
		*font = (*font)[1:]
		*bitsLeft = 32
	}
}

func SubPixmap(src Pixmap, x0, y0, x1, y1 int) Pixmap {
	if x0 > x1 {
		panic("x0 > x1")
	}
	if y0 > y1 {
		panic("y0 > y1")
	}
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > src.Width {
		x1 = src.Width
	}
	if y1 > src.Height {
		y1 = src.Height
	}

	var out Pixmap
	out.Pixels = src.Pixels[x0+y0*src.Stride:]
	out.Width = x1 - x0
	out.Height = y1 - y0
	out.Stride = src.Stride
	out.Alpha = src.Alpha
	return out
}

func SubPixmapWH(src Pixmap, x, y, width, height int) Pixmap {
	return SubPixmap(src, x, y, x+width, y+height)
}
