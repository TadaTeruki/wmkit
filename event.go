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
	eventType 		EventType
	target_xwindow	C.xcb_window_t
	available 		bool
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
	event.target_xwindow = 0
	sc.sendEvent(&event)
}

func (sc *Screen) Main(mainloop func(event *Event)){
	for {
		event := sc.nextEvent()
		
		if event.available == false { continue }
		if event.eventType == QuitNotify { break }

		mainloop(event)

		switch event.eventType {
			case QuitRequest : {
				var new_event Event
				new_event.eventType = QuitNotify
				new_event.available = true
				new_event.target_xwindow = 0
				sc.sendEvent(&new_event)
			}
			default: {
				
			}
		}
	}
}

func (sc *Screen) nextEvent() *Event{
	var event Event
	event.available 	 = true
	event.target_xwindow = 0
	event.screen		 = sc

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

			button_event := (*C.xcb_button_press_event_t)(unsafe.Pointer(generic_event))
			event.eventType = ButtonPressNotify
			event.target_xwindow = button_event.event
			
		}

		case C.XCB_BUTTON_RELEASE:{
			log.Println("wmkit : listen XCB_BUTTON_RELEASE")
			event.eventType = ButtonReleaseNotify
		}

		case C.XCB_EXPOSE:{
			log.Println("wmkit : listen XCB_EXPOSE")

			expose_event := (*C.xcb_expose_event_t)(unsafe.Pointer(generic_event))
			event.eventType = ExposeNotify
			event.target_xwindow = expose_event.window
		}

		default:{
			event.eventType = None
		}

	}
	
	C.free(unsafe.Pointer(generic_event))

	return &event

}
