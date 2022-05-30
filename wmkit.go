package wmkit

/*
#cgo pkg-config: xcb
#include <xcb/xcb.h>
*/
import "C"

import (
	"os"
	"unsafe"
)

type Screen struct {
	connection 	*C.xcb_connection_t
	xscreen		*C.xcb_screen_t
	panels		[]Panel
	rootPanel	Panel
	logFile   	*os.File
	eventQueue	*eventQueue
	noevent		bool
}

type XY struct {
	X, Y int
}

type WH struct {
	W, H uint
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
	sc.panels		= []Panel{}
	
	sc.rootPanel 	= sc.initialPanel()
	sc.rootPanel.xwindow = sc.xscreen.root
	
	values := [1]C.uint32_t{ C.XCB_EVENT_MASK_SUBSTRUCTURE_REDIRECT | C.XCB_EVENT_MASK_SUBSTRUCTURE_NOTIFY }
	C.xcb_change_window_attributes(
		sc.connection, sc.xscreen.root, C.XCB_CW_EVENT_MASK, unsafe.Pointer(&values[0]));
	
}

func (sc *Screen) GetRootPanel() *Panel{
	return &sc.rootPanel
}

func (sc *Screen) Disconnect() {
	for _, panel := range sc.panels {
		if panel.xwindow == sc.xscreen.root { continue }
		panel.internalDestroy()
	}
	C.xcb_disconnect(sc.connection)
}

func (sc *Screen) Flush(){
	C.xcb_flush(sc.connection);
}

