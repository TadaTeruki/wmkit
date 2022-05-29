package wmkit

/*
#cgo pkg-config: xcb
#include <xcb/xcb.h>
*/
import "C"

import (
	//"log"
	//"unsafe"
)

type EventType int

const (
	EventType_None			EventType =  0
	QuitRequest 			EventType = -1
	quitNotify  			EventType =  1
	ExposeNotify  			EventType =  2
	ButtonPressNotify 		EventType =  3
	ButtonReleaseNotify 	EventType =  4
	ButtonMotionNotify 		EventType =  5
	PointerMotionNotify 	EventType =  6
)

type Event struct {
	eventType 				EventType
	targetXwindow			C.xcb_window_t
	requestIsAvailable 		bool
	screen					*Screen
	buttonProperty			*EventButtonProperty
	motionProperty			*EventMotionProperty
}

type EventButtonProperty struct {
	EventX	int
	EventY	int
	RootX	int
	RootY	int
	Detail 	uint
}

type EventMotionProperty struct {
	EventX	int
	EventY	int
	RootX	int
	RootY	int
}

type eventQueue struct {
	event Event
	next *eventQueue
}

func (event *Event) GetType() EventType {
	return event.eventType
}

func (event *Event) GetPanel() *Panel {
	for _, panel := range event.screen.panels {
		if panel.xwindow == event.targetXwindow {
			return &panel
		}
	}
	return nil
}

func (event *Event) GetButtonProperty() *EventButtonProperty {
	return event.buttonProperty
}

func (event *Event) GetMotionProperty() *EventMotionProperty {
	return event.motionProperty
}

func (event *Event) RejectRequest(){
	event.requestIsAvailable = false
}

