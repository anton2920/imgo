//go:build unix

package imgo

/*
#cgo CFLAGS: -I/usr/local/include
#cgo LDFLAGS: -L/usr/local/lib
#cgo LDFLAGS: -lX11 -lm -lxcb -lXau -lXdmcp

#include <X11/Xlib.h>
*/
import "C"
import (
	"errors"
	"runtime"
	"unsafe"
)

type RGBA struct {
	a, r, g, b byte
}

type platformWindow struct {
	wmDeleteWindow C.Atom
	display        *C.Display
	window         C.Window
	root           C.Window
	visual         *C.Visual
	screen         C.int
	gc             C.GC
}

func platformGetEvent(w *Window) Event {
	var event C.XEvent
	C.XNextEvent(w.display, &event)

	/* NOTE(anton2920): convoluted way of saying 'event.type'. */
	switch *(*C.int)(unsafe.Pointer(&event)) {
	case C.ClientMessage:
		clientEvent := *(*C.XClientMessageEvent)(unsafe.Pointer(&event))
		data := *(*C.int)(unsafe.Pointer(&clientEvent.data[0]))

		if C.Atom(data) == w.wmDeleteWindow {
			return DestroyEvent{}
		}
	case C.Expose:
		return PaintEvent{}
	case C.ConfigureNotify:
		configureEvent := *(*C.XConfigureEvent)(unsafe.Pointer(&event))
		eventWidth := int(configureEvent.width)
		eventHeight := int(configureEvent.height)

		if (eventWidth != w.width) || (eventHeight != w.height) {
			return ResizeEvent{Width: eventWidth, Height: eventHeight}
		}
	case C.ButtonPress:
		buttonEvent := *(*C.XButtonEvent)(unsafe.Pointer(&event))
		eventX := int(buttonEvent.x)
		eventY := int(buttonEvent.y)

		return MouseButtonDownEvent{X: eventX, Y: eventY, Button: MouseButton(buttonEvent.button)}
	case C.ButtonRelease:
		buttonEvent := *(*C.XButtonEvent)(unsafe.Pointer(&event))
		eventX := int(buttonEvent.x)
		eventY := int(buttonEvent.y)

		return MouseButtonUpEvent{X: eventX, Y: eventY, Button: MouseButton(buttonEvent.button)}
	case C.MotionNotify:
		motionEvent := *(*C.XMotionEvent)(unsafe.Pointer(&event))
		eventX := int(motionEvent.x)
		eventY := int(motionEvent.y)

		return MouseMoveEvent{X: eventX, Y: eventY}
	}

	return nil
}

/* PlatformDrawPixmap draws pixels of certain width and height to a certain (x,y) screen position. */
func platformDrawPixmap(w *Window, x, y int, pixels []RGBA, width, height int, stride int) {
	var pinner runtime.Pinner
	var image C.XImage

	pinner.Pin(unsafe.SliceData(pixels))

	image.width = C.int(width)
	image.height = C.int(height)
	image.format = C.ZPixmap
	image.data = (*C.char)(unsafe.Pointer(unsafe.SliceData(pixels)))
	image.bitmap_unit = C.int(unsafe.Sizeof(pixels[0]) * 8)
	image.bitmap_pad = C.int(unsafe.Sizeof(pixels[0]) * 8)
	image.depth = 24
	image.bytes_per_line = C.int(width * int(unsafe.Sizeof(pixels[0])))
	image.bits_per_pixel = C.int(unsafe.Sizeof(pixels[0]) * 8)
	image.red_mask = w.visual.red_mask
	image.green_mask = w.visual.green_mask
	image.blue_mask = w.visual.blue_mask
	C.XInitImage(&image)

	C.XPutImage(w.display, w.window, w.gc, &image, 0, 0, C.int(x), C.int(y), C.uint(width), C.uint(height))

	pinner.Unpin()
}

func platformHasEvents(w *Window) bool {
	return C.XPending(w.display) > 0
}

func platformMouseCapture(w *Window, capture bool) {
	if capture {
		C.XGrabPointer(w.display, w.window, 0, C.ButtonMotionMask|C.ButtonPressMask|C.ButtonReleaseMask, C.GrabModeAsync, C.GrabModeAsync, w.root, C.None, C.CurrentTime)
	} else {
		C.XUngrabPointer(w.display, C.CurrentTime)
	}
}

func platformNewWindow(w *Window) error {
	w.display = C.XOpenDisplay(nil)
	if w.display == nil {
		return errors.New("failed to open display")
	}

	w.screen = C.XDefaultScreen(w.display)
	w.visual = C.XDefaultVisual(w.display, w.screen)
	if w.visual.class != C.TrueColor {
		return errors.New("cannot handle non-true color visual")
	}
	w.root = C.XDefaultRootWindow(w.display)
	w.gc = C.XDefaultGC(w.display, w.screen)

	w.window = C.XCreateSimpleWindow(w.display, w.root, C.int(w.x), C.int(w.y), C.uint(w.width), C.uint(w.height), 1, 0, 0)
	C.XSelectInput(w.display, w.window, C.ExposureMask|C.KeyPressMask|C.KeyReleaseMask|C.ButtonPressMask|C.ButtonReleaseMask|C.PointerMotionMask|C.StructureNotifyMask)

	C.XStoreName(w.display, w.window, C.CString(w.title))

	if (w.flags & WindowHidden) == 0 {
		C.XMapWindow(w.display, w.window)
	}

	w.wmDeleteWindow = C.XInternAtom(w.display, C.CString("WM_DELETE_WINDOW"), 1)
	C.XSetWMProtocols(w.display, w.window, &w.wmDeleteWindow, 1)

	return nil
}

func platformRedrawAll(w *Window) {
	C.XClearArea(w.display, w.window, 0, 0, 1, 1, C.int(1))
	C.XFlush(w.display)
}

func platformWindowClose(w *Window) {
	C.XCloseDisplay(w.display)
}
