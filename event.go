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

type EventType int

const (
	EventType_None		EventType =  0
	QuitRequest 		EventType = -1
	quitNotify  		EventType =  1
	ExposeNotify  		EventType =  2
	ButtonPressNotify 	EventType =  3
	ButtonReleaseNotify EventType =  4
)

type Event struct {
	eventType 		EventType
	target_xwindow	C.xcb_window_t
	request_available 		bool
	screen			*Screen
}

type EventQueue struct {
	event Event
	next *EventQueue
}

func (event *Event) GetType() EventType {
	return event.eventType
}

func (event *Event) GetPanel() *Panel {
	for _, panel := range event.screen.panels {
		if panel.xwindow == event.target_xwindow {
			return &panel
		}
	}
	return nil
}

func (event *Event) RejectRequest(){
	event.request_available = false
}

func (sc *Screen) sendEvent(event *Event) {
	new_eq := &EventQueue{ *event, nil }

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
	event.request_available = true
	event.target_xwindow = 0
	sc.sendEvent(&event)
}

func (sc *Screen) CatchRequest(event *Event) {

	if event.eventType == quitNotify { 
		sc.noevent = true
		return
	}

	if event.request_available == false { return }

	switch event.eventType {
		case QuitRequest : {
			var new_event Event
			new_event.eventType = quitNotify
			new_event.request_available = false
			new_event.target_xwindow = 0
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
	event.request_available		= false
	event.target_xwindow 		= 0
	event.screen		 		= sc

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
			event.target_xwindow = button_event.event
			
		}

		case C.XCB_BUTTON_RELEASE:{
			log.Println("wmkit : listen XCB_BUTTON_RELEASE")
			button_event := (*C.xcb_button_release_event_t)(unsafe.Pointer(generic_event))
			event.eventType = ButtonReleaseNotify
			event.target_xwindow = button_event.event
		}

		case C.XCB_EXPOSE:{
			log.Println("wmkit : listen XCB_EXPOSE")

			expose_event := (*C.xcb_expose_event_t)(unsafe.Pointer(generic_event))
			event.eventType = ExposeNotify
			event.target_xwindow = expose_event.window
		}

		default:{
			event.eventType = EventType_None
		}

	}
	
	

	return &event

}
