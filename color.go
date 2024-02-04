package imgo

type Color uint32

var (
	ColorBlack = ColorRGB(0, 0, 0)
	ColorWhite = ColorRGB(255, 255, 255)
)

func (c Color) AlphaComponent() byte {
	return byte((c >> 24) & 0xFF)
}

func (c Color) BlueComponent() byte {
	return byte(c & 0xFF)
}

func (c Color) GreenComponent() byte {
	return byte((c >> 8) & 0xFF)
}

func (c Color) Invisible() bool {
	return c <= 0x00FFFFFF
}

/* Opaque returns true if color is 100% non-transparent. */
func (c Color) Opaque() bool {
	return c >= 0xFF000000
}

func (c Color) RedComponent() byte {
	return byte((c >> 16) & 0xFF)
}

func ColorAverage(c, d Color) Color {
	c = (c >> 1) & 0x7f7f7f7f
	d = (d >> 1) & 0x7f7f7f7f
	return c + d + (c & d & 0x01010101)
}

func ColorDark(c Color) Color {
	return ColorAverage(c, ColorBlack)
}

func ColorGrey(c Color) Color {
	return ColorRGB(byte(c), byte(c), byte(c))
}

func ColorLite(c Color) Color {
	return ColorAverage(c, ColorWhite)
}

func ColorRGB(r, g, b byte) Color {
	return ColorRGBA(r, g, b, 255)
}

func ColorRGBA(r, g, b, a byte) Color {
	return Color(a)<<24 | Color(r)<<16 | Color(g)<<8 | Color(b)
}
