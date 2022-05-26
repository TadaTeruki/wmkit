package wmkit

/*
#cgo pkg-config: xcb cairo
#include <xcb/xcb.h>
#include <cairo.h>
#include <cairo-xcb.h>
#include <stdlib.h>
*/
import "C"

import "unsafe"
import "fmt"

type EventType uint

const (
	None		EventType = 0
	QuitRequest EventType = 1
	QuitNotify  EventType = 2
	DrawRequest EventType = 3
	DrawNotify  EventType = 4
)

type Event struct {
	Type EventType
}

func (sc *Screen) NextEvent() *Event{
	var event Event

	generic_event := C.xcb_wait_for_event(sc.connection)

	if generic_event == nil {
		event.Type = None
		return &event
	}
	
	switch(generic_event.response_type){

		case C.XCB_BUTTON_PRESS:{
		}

		default:{
		}

	}
	

	C.free(unsafe.Pointer(generic_event))

	return &event

}
