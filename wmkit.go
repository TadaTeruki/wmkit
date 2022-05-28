package wmkit

/*
#cgo pkg-config: xcb cairo
#include <xcb/xcb.h>
#include <cairo.h>
#include <cairo-xcb.h>
*/
import "C"

import (
	"os"
)

type Screen struct {
	connection 	*C.xcb_connection_t
	xscreen		*C.xcb_screen_t
	panels		[]Panel
	logFile   	*os.File
	eventQueue	*EventQueue
}

type XYWH struct {
	X, Y, W, H int
}

func (sc *Screen) Connect() {
	sc.connection	= C.xcb_connect(nil, nil)
	sc.xscreen		= C.xcb_setup_roots_iterator(C.xcb_get_setup(sc.connection)).data
	sc.eventQueue 	= nil
}

func (sc *Screen) Disconnect() {
	C.xcb_disconnect(sc.connection)
}

func (sc *Screen) Flush(){
	C.xcb_flush(sc.connection);
}

