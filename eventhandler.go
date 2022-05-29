package wmkit
/*
#cgo pkg-config: xcb
#include <xcb/xcb.h>
#include <stdlib.h>
*/
import "C"

import (
	"log"
	"unsafe"
)

func (sc *Screen) sendEvent(event *Event) {
	new_eq := &eventQueue{ *event, nil }

	if sc.eventQueue == nil {
		sc.eventQueue = new_eq
		return
	} else {
		last_eq := sc.eventQueue
		for ; last_eq.next != nil; last_eq = last_eq.next { }
		last_eq.next = new_eq
	}
}

func (sc *Screen) RequestQuit() {
	var event Event
	event.eventType = QuitRequest
	event.requestIsAvailable = true
	event.targetXwindow = 0
	sc.sendEvent(&event)
}

func (sc *Screen) CatchRequest(event *Event) {

	if event.eventType == quitNotify { 
		sc.noevent = true
		return
	}

	if event.requestIsAvailable == false { return }

	switch event.eventType {
		case QuitRequest : {
			var new_event Event
			new_event.eventType = quitNotify
			new_event.requestIsAvailable = false
			new_event.targetXwindow = 0
			sc.sendEvent(&new_event)
		}
		default: {
			
		}
	}
}

func (sc *Screen) NextEvent() *Event{

	if sc.noevent == true {
		return nil
	}

	var event Event
	event.requestIsAvailable	= false
	event.targetXwindow 		= 0
	event.screen		 		= sc
	event.buttonProperty		= nil
	event.motionProperty		= nil

	if sc.eventQueue != nil {
		event = sc.eventQueue.event
		sc.eventQueue = sc.eventQueue.next
		return &event
	}

	generic_event := C.xcb_wait_for_event(sc.connection)
	defer C.free(unsafe.Pointer(generic_event))

	if generic_event == nil {
		event.eventType = EventType_None
		return &event
	}
	
	switch generic_event.response_type {

		case C.XCB_BUTTON_PRESS:{
			log.Println("wmkit : listen XCB_BUTTON_PRESS")

			button_event := (*C.xcb_button_press_event_t)(unsafe.Pointer(generic_event))
			event.eventType = ButtonPressNotify
			event.targetXwindow = button_event.event
			event.buttonProperty = &EventButtonProperty{
				int(button_event.event_x), int(button_event.event_y), int(button_event.root_x), int(button_event.root_y), uint(button_event.detail),
			}
			
		}

		case C.XCB_BUTTON_RELEASE:{
			log.Println("wmkit : listen XCB_BUTTON_RELEASE")
			button_event := (*C.xcb_button_release_event_t)(unsafe.Pointer(generic_event))
			event.eventType = ButtonReleaseNotify
			event.targetXwindow = button_event.event
			event.buttonProperty = &EventButtonProperty{
				int(button_event.event_x), int(button_event.event_y), int(button_event.root_x), int(button_event.root_y), uint(button_event.detail),
			}

		}

		case C.XCB_EXPOSE:{
			log.Println("wmkit : listen XCB_EXPOSE")

			expose_event := (*C.xcb_expose_event_t)(unsafe.Pointer(generic_event))
			event.eventType = ExposeNotify
			event.targetXwindow = expose_event.window
		}

		case C.XCB_MOTION_NOTIFY:{
			//log.Println("wmkit : listen XCB_MOTION")

			motion_event := (*C.xcb_motion_notify_event_t)(unsafe.Pointer(generic_event))
			event.eventType = ButtonMotionNotify
			event.targetXwindow = motion_event.event

			event.motionProperty = &EventMotionProperty{
				int(motion_event.event_x), int(motion_event.event_y), int(motion_event.root_x), int(motion_event.root_y), 
			}
		}

		default:{
			event.eventType = EventType_None
		}

	}
	
	return &event
}
