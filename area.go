package wmkit

/*
#cgo pkg-config: xcb cairo
#include <xcb/xcb.h>
#include <cairo.h>
#include <cairo-xcb.h>
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
	surface			*cairo.Surface
}

type Panel struct {
	panelType 		PanelType
	xwindow  		C.xcb_window_t
	drawComponent 	*panelDrawComponent
}

func initialPanel() Panel {
	var panel Panel
	panel.panelType 			= PanelType_None
	panel.xwindow 				= 0
	panel.drawComponent			= nil
	return panel
}

func (sc *Screen) newXWindow(panel *Panel, xywh XYWH) {

	panel.xwindow = C.xcb_generate_id(sc.connection)
	values := [2]C.uint32_t{ C.uint32_t(1), C.XCB_EVENT_MASK_EXPOSURE | C.XCB_EVENT_MASK_BUTTON_PRESS | C.XCB_EVENT_MASK_BUTTON_RELEASE }

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
}

func (sc *Screen) NewPanel(panelType PanelType, xywh XYWH) *Panel{

	sc.panels = append(sc.panels, initialPanel())
	panel := &sc.panels[len(sc.panels)-1]

	panel.panelType 			= panelType
	panel.xwindow 				= 0
	panel.drawComponent			= nil

	if panel.panelType == PanelType_None { return panel }

	sc.newXWindow(panel, xywh)

	switch panel.panelType {

	case PanelType_Plain:{
		break
	}
	case PanelType_Drawable:{
		panel.drawComponent = &panelDrawComponent{ nil }

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
			cairo_surface := C.cairo_xcb_surface_create(sc.connection, panel.xwindow, visual_type, C.int(xywh.W), C.int(xywh.H))
			cairo_context := C.cairo_create(cairo_surface)
			panel.drawComponent.surface = cairo.NewSurfaceFromC(
				cairo.Cairo_surface(unsafe.Pointer(cairo_surface)),
				cairo.Cairo_context(unsafe.Pointer(cairo_context)),
			)
	
		}
	}
	default:break
	}

	

	return panel
}

func (sc *Screen) Map(panel *Panel){
	C.xcb_map_window(sc.connection, panel.xwindow)
}

func (sc *Screen) Destroy(panel *Panel){
	C.xcb_unmap_window(sc.connection,   panel.xwindow)
	C.xcb_destroy_window(sc.connection, panel.xwindow)
	if panel.panelType == PanelType_Drawable {
		panel.drawComponent.surface.Destroy()
	}
}

func (panel *Panel) GetCairoSurface() *cairo.Surface {
	return panel.drawComponent.surface
}
