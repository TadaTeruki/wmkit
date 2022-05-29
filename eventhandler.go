package wmkit
/*
#cgo CFLAGS: -Wall
#cgo pkg-config: xcb xcb-util

#include <xcb/xcb.h>
#include <xcb/xcb_util.h>
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
		case MapRequest : {
			C.xcb_map_window(sc.connection, event.targetXwindow)
			sc.Flush()
		}
		case ConfigureRequest : {
			values := [4]C.uint32_t{
				C.uint(event.configureProperty.X),
				C.uint(event.configureProperty.Y),
				C.uint(event.configureProperty.W),
				C.uint(event.configureProperty.H),
			}

			C.xcb_configure_window(sc.connection, event.targetXwindow,
				C.XCB_CONFIG_WINDOW_X | C.XCB_CONFIG_WINDOW_Y | C.XCB_CONFIG_WINDOW_WIDTH | C.XCB_CONFIG_WINDOW_HEIGHT,
				unsafe.Pointer(&values[0]))
			sc.Flush()
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
	event.configureProperty		= nil

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
			button_event := (*C.xcb_button_press_event_t)(unsafe.Pointer(generic_event))
			event.eventType = ButtonPressNotify
			event.targetXwindow = button_event.event
			event.buttonProperty = &EventButtonProperty{
				int(button_event.event_x), int(button_event.event_y), int(button_event.root_x), int(button_event.root_y), uint(button_event.detail),
			}
			
		}

		case C.XCB_BUTTON_RELEASE:{
			button_event := (*C.xcb_button_release_event_t)(unsafe.Pointer(generic_event))
			event.eventType = ButtonReleaseNotify
			event.targetXwindow = button_event.event
			event.buttonProperty = &EventButtonProperty{
				int(button_event.event_x), int(button_event.event_y), int(button_event.root_x), int(button_event.root_y), uint(button_event.detail),
			}

		}
		
		case C.XCB_EXPOSE:{
			expose_event := (*C.xcb_expose_event_t)(unsafe.Pointer(generic_event))
			event.eventType = ExposeNotify
			event.targetXwindow = expose_event.window
		}

		case C.XCB_MOTION_NOTIFY:{
			motion_event := (*C.xcb_motion_notify_event_t)(unsafe.Pointer(generic_event))
			event.eventType = ButtonMotionNotify
			event.targetXwindow = motion_event.event

			event.motionProperty = &EventMotionProperty{
				int(motion_event.event_x), int(motion_event.event_y), int(motion_event.root_x), int(motion_event.root_y), 
			}
		}

		case C.XCB_MAP_NOTIFY:{
			event.eventType = MapNotify
			map_event := (*C.xcb_map_notify_event_t)(unsafe.Pointer(generic_event))
			event.targetXwindow = map_event.window
		}

		case C.XCB_MAP_REQUEST:{
			event.eventType = MapRequest
			map_event := (*C.xcb_map_request_event_t)(unsafe.Pointer(generic_event))
			event.targetXwindow = map_event.window
			event.requestIsAvailable = true
		}

		case C.XCB_CONFIGURE_REQUEST:{
			event.eventType = ConfigureRequest
			configure_event := (*C.xcb_configure_request_event_t)(unsafe.Pointer(generic_event))
			event.targetXwindow = configure_event.window
			event.configureProperty = &EventConfigureProperty{
				int(configure_event.x), int(configure_event.y), uint(configure_event.width), uint(configure_event.height), 
			}
			event.requestIsAvailable = true
		}

		case C.XCB_CREATE_NOTIFY:{
			event.eventType = CreateNotify
			create_event := (*C.xcb_create_notify_event_t)(unsafe.Pointer(generic_event))
			event.targetXwindow = create_event.window
		}

		case C.XCB_CONFIGURE_NOTIFY:{
			event.eventType = ConfigureNotify
			configure_event := (*C.xcb_configure_notify_event_t)(unsafe.Pointer(generic_event))
			event.targetXwindow = configure_event.window
		}

		// can't use yet
		case C.XCB_UNMAP_NOTIFY:{
			unmap_event := (*C.xcb_unmap_notify_event_t)(unsafe.Pointer(generic_event))
			event.targetXwindow = unmap_event.window
		}

		// can't use yet
		case C.XCB_DESTROY_NOTIFY:{
			destroy_event := (*C.xcb_destroy_notify_event_t)(unsafe.Pointer(generic_event))
			event.targetXwindow = destroy_event.window
		}

		default:{
			event.eventType = EventType_None
		}

	}

	log.Println("wmkit : listen", C.GoString(C.xcb_event_get_label(generic_event.response_type)), event.targetXwindow)
	
	return &event
}
