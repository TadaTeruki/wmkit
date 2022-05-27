package wmkit

/*
#cgo pkg-config: xcb cairo
#include <xcb/xcb.h>
#include <cairo.h>
#include <cairo-xcb.h>
#include <stdlib.h>
*/
import "C"

import (
	"log"
	"unsafe"
)

type EventType int

const (
	None				EventType =  0
	QuitRequest 		EventType = -1
	QuitNotify  		EventType =  1
	ExposeNotify  		EventType =  2
	ButtonPressNotify 	EventType =  3
	ButtonReleaseNotify EventType =  4
)

type Event struct {
	eventType EventType
	available bool
}

type EventQueue struct {
	event Event
	next *EventQueue
}

func (event *Event) GetType() EventType {
	return event.eventType
}

func (event *Event) RejectRequest(){
	event.available = false
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
	event.available = true
	sc.sendEvent(&event)
}

func (sc *Screen) Main(mainloop func(event *Event)){
	for {
		event := sc.nextEvent()
		mainloop(event)
		if event.available == false { continue }
		if event.eventType == QuitRequest { break }

		switch event.eventType {
			default: {
				
			}
		}

		
	}
}

func (sc *Screen) nextEvent() *Event{
	var event Event
	event.available = true

	if sc.eventQueue != nil {
		event = sc.eventQueue.event
		sc.eventQueue = sc.eventQueue.next
		return &event
	}

	generic_event := C.xcb_wait_for_event(sc.connection)

	if generic_event == nil {
		event.eventType = None
		return &event
	}
	
	switch generic_event.response_type {

		case C.XCB_BUTTON_PRESS:{
			log.Println("wmkit : listen XCB_BUTTON_PRESS")
			event.eventType = ButtonPressNotify
		}

		case C.XCB_BUTTON_RELEASE:{
			log.Println("wmkit : listen XCB_BUTTON_RELEASE")
			event.eventType = ButtonReleaseNotify
		}

		case C.XCB_EXPOSE:{
			log.Println("wmkit : listen XCB_EXPOSE")
			event.eventType = ExposeNotify
		}

		default:{
			event.eventType = None
		}

	}
	

	C.free(unsafe.Pointer(generic_event))

	return &event

}
