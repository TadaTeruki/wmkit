package wmkit

/*
#cgo pkg-config: xcb cairo
#include <xcb/xcb.h>
#include <cairo.h>
#include <cairo-xcb.h>
#include <stdlib.h>
*/
import "C"

import(
	"unsafe"
	cairo "github.com/ungerik/go-cairo"
)

// === Window === //

type PanelType int

const(
	PanelType_None	 		PanelType = 0
	PanelType_Plain 		PanelType = 1
	PanelType_Drawable		PanelType = 2
)

type panelDrawComponent struct {
	c_surface *C.cairo_surface_t
	c_context *C.cairo_t
}

type Panel struct {
	panelType 		PanelType
	xwindow  		C.xcb_window_t
	drawComponent 	*panelDrawComponent
	screen			*Screen
}

func (sc *Screen) initialPanel() Panel {
	var panel Panel
	panel.panelType 			= PanelType_None
	panel.xwindow 				= 0
	panel.drawComponent			= nil
	panel.screen				= sc
	return panel
}

func (sc *Screen) initPanelWithXWindow(panel *Panel, xwindow *C.xcb_window_t, xywh *XYWH, allowedEventType []EventType, overrideRedirect bool) {


	if xwindow == nil {
		panel.xwindow = C.xcb_generate_id(sc.connection)
	} else {
		panel.xwindow = *xwindow
	}

	var eventMask C.uint32_t = 0

	for _, eventType := range allowedEventType {
		switch eventType {
		case EventType_None			: break
		case ExposeNotify			: eventMask |= C.XCB_EVENT_MASK_EXPOSURE
		case ButtonPressNotify		: eventMask |= C.XCB_EVENT_MASK_BUTTON_PRESS
		case ButtonReleaseNotify	: eventMask |= C.XCB_EVENT_MASK_BUTTON_RELEASE
		case ButtonMotionNotify		: eventMask |= C.XCB_EVENT_MASK_BUTTON_MOTION
		case PointerMotionNotify	: eventMask |= C.XCB_EVENT_MASK_POINTER_MOTION
		default:break
		}
	}

	var overrideRedirectFlag C.uint32_t = 0
	if overrideRedirect { overrideRedirectFlag = 1 }

	values := [2]C.uint32_t{ overrideRedirectFlag, eventMask }

	if xwindow == nil {
		C.xcb_create_window (sc.connection,
			C.XCB_COPY_FROM_PARENT,
			panel.xwindow,
			sc.xscreen.root,
			C.short(xywh.X),  C.short(xywh.Y),
			C.ushort(xywh.W), C.ushort(xywh.H),
			0,
			C.XCB_WINDOW_CLASS_INPUT_OUTPUT,
			sc.xscreen.root_visual,
			C.XCB_CW_OVERRIDE_REDIRECT | C.XCB_CW_EVENT_MASK ,
			unsafe.Pointer(&values[0]))
	} else {
		C.xcb_change_window_attributes(
			sc.connection, panel.xwindow, 
			C.XCB_CW_OVERRIDE_REDIRECT | C.XCB_CW_EVENT_MASK,
			unsafe.Pointer(&values[0]));
	}

}

func (panel *Panel) applyPanelType(){

	sc := panel.screen

	xywh := panel.GetXYWH()

	panel.drawComponent	= nil

	switch panel.panelType {

		case PanelType_Plain:{
		}
		case PanelType_Drawable:{
			panel.drawComponent = &panelDrawComponent{ nil, nil }
	
			visual_is_found := false
			var visual_type *C.xcb_visualtype_t
		
		
			depth_itr := C.xcb_screen_allowed_depths_iterator(sc.xscreen)
			for ; visual_is_found == false && depth_itr.rem != 0; C.xcb_depth_next(&depth_itr) {
		
				visual_itr := C.xcb_depth_visuals_iterator(depth_itr.data)
				for ; visual_is_found == false && visual_itr.rem != 0; C.xcb_visualtype_next(&visual_itr) {
					if (sc.xscreen.root_visual == visual_itr.data.visual_id) {
						visual_type = visual_itr.data
						visual_is_found = true
					}
				}
			}
		
			if visual_type != nil {
				panel.drawComponent.c_surface = C.cairo_xcb_surface_create(sc.connection, panel.xwindow, visual_type, C.int(xywh.W), C.int(xywh.H))
				panel.drawComponent.c_context = C.cairo_create(panel.drawComponent.c_surface)
			}
		}
		default:break
	}
}

func (sc *Screen) NewPanel(panelType PanelType, xywh XYWH, allowedEventType []EventType, overrideRedirect bool) *Panel{

	sc.panels = append(sc.panels, sc.initialPanel())
	panel := &sc.panels[len(sc.panels)-1]

	panel.panelType 			= panelType
	panel.xwindow 				= 0

	if panel.panelType == PanelType_None { return panel }

	sc.initPanelWithXWindow(panel, nil, &xywh, allowedEventType, overrideRedirect)

	panel.applyPanelType()

	return panel
}
/*
func (panel *Panel) changePanelType(panelType PanelType){

	if panel.panelType == panelType {
		return
	}

	panel.panelType = panelType
	panel.applyPanelType()	
}
*/

func (panel *Panel) Map(){
	sc := panel.screen
	C.xcb_map_window(sc.connection, panel.xwindow)
}

func (panel *Panel) Unmap(){
	sc := panel.screen
	C.xcb_unmap_window(sc.connection,   panel.xwindow)
}

func (panel *Panel) internalDestroy(){
	sc := panel.screen
	C.xcb_destroy_window(sc.connection, panel.xwindow)
	if panel.panelType == PanelType_Drawable {
		C.cairo_destroy(panel.drawComponent.c_context)
		C.cairo_surface_destroy(panel.drawComponent.c_surface)
	}
}

func (panel *Panel) Destroy(){
	sc := panel.screen
	panel_num := len(sc.panels)

	for i, p := range sc.panels {
		if panel.xwindow == p.xwindow {
			sc.panels[i] = sc.panels[panel_num-1]
			sc.panels = sc.panels[:panel_num-1]
		}
	}
	panel.internalDestroy()
}

func (panel *Panel) GetCairoSurface() *cairo.Surface {
	if panel.drawComponent == nil {
		return nil
	}
	return cairo.NewSurfaceFromC(
		cairo.Cairo_surface(unsafe.Pointer(panel.drawComponent.c_surface)),
		cairo.Cairo_context(unsafe.Pointer(panel.drawComponent.c_context)),
	)
}

func (panel *Panel) GetXYWH() XYWH{
	sc := panel.screen
	geom := C.xcb_get_geometry_reply(sc.connection, C.xcb_get_geometry(sc.connection, panel.xwindow), nil)
	defer C.free(unsafe.Pointer(geom))

	if geom == nil {
		return XYWH{0, 0, 0, 0}
	}

	xywh := XYWH{int(geom.x), int(geom.y), uint(geom.width), uint(geom.height)}

	return xywh
}

func (panel *Panel) ApplyXYWH(xywh XYWH){
	sc := panel.screen
	values := [4]C.uint32_t{ C.uint(xywh.X), C.uint(xywh.Y), C.uint(xywh.W), C.uint(xywh.H) }
	C.xcb_configure_window (sc.connection, panel.xwindow,
		C.XCB_CONFIG_WINDOW_X | C.XCB_CONFIG_WINDOW_Y | C.XCB_CONFIG_WINDOW_WIDTH | C.XCB_CONFIG_WINDOW_HEIGHT, unsafe.Pointer(&values[0]))

	if panel.panelType == PanelType_Drawable && panel.drawComponent != nil && panel.drawComponent.c_surface != nil{
		C.cairo_xcb_surface_set_size(panel.drawComponent.c_surface, C.int(xywh.W), C.int(xywh.H) )
	}
}

