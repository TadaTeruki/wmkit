package wmkit

/*
#cgo pkg-config: xcb
#include <xcb/xcb.h>
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
	noevent		bool
}

type XYWH struct {
	X, Y int
	W, H uint
}

func (sc *Screen) Connect() {
	sc.connection	= C.xcb_connect(nil, nil)
	sc.xscreen		= C.xcb_setup_roots_iterator(C.xcb_get_setup(sc.connection)).data
	sc.eventQueue 	= nil
	sc.noevent		= false
}

func (sc *Screen) Disconnect() {
	for _, panel := range sc.panels {
		panel.internalDestroy()
	}
	C.xcb_disconnect(sc.connection)
}

func (sc *Screen) Flush(){
	C.xcb_flush(sc.connection);
}

