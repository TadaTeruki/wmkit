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
	//cairo "github.com/ungerik/go-cairo"
)

// === Area === //

type Area interface {
	getXWindow() C.xcb_window_t
	setXWindow(C.xcb_window_t)
	register(*Screen)
	destroy()
}

func (sc *Screen) initArea(area Area, xywh XYWH) {

	xwindow := C.xcb_generate_id(sc.connection)

	/*
	uint32_t             mask = 0;
	uint32_t             values[2];
	mask = XCB_GC_FOREGROUND | XCB_GC_GRAPHICS_EXPOSURES;
	values[0] = screen->black_pixel;
	values[1] = 0;
	*/
	//valwin[0] = XCB_EVENT_MASK_EXPOSURE | XCB_EVENT_MASK_BUTTON_PRESS;

	values := []C.uint32_t{ C.XCB_EVENT_MASK_EXPOSURE | C.XCB_EVENT_MASK_BUTTON_PRESS }

	C.xcb_create_window (sc.connection,
		C.XCB_COPY_FROM_PARENT,
	 	xwindow,
		sc.xscreen.root,
		C.short(xywh.X),  C.short(xywh.Y),
		C.ushort(xywh.W), C.ushort(xywh.H),
		10,
		C.XCB_WINDOW_CLASS_INPUT_OUTPUT,
		sc.xscreen.root_visual,
		C.XCB_CW_EVENT_MASK,
		unsafe.Pointer(&values[0]))
	
	area.setXWindow(xwindow)
	area.register(sc)
}

func (sc *Screen) Map(area Area){
	C.xcb_map_window(sc.connection, area.getXWindow())
}

func (sc *Screen) Destroy(area Area){
	C.xcb_unmap_window(sc.connection,   area.getXWindow())
	C.xcb_destroy_window(sc.connection, area.getXWindow())
	area.destroy()
}



// --- PlainArea --- //

type PlainArea struct {
	xwindow C.xcb_window_t
}

func (p_area *PlainArea) getXWindow() C.xcb_window_t {
	return p_area.xwindow
}

func (p_area *PlainArea) setXWindow(xwindow C.xcb_window_t) {
	p_area.xwindow = xwindow
}

func (p_area *PlainArea) register(sc *Screen){
	sc.pAreas = append(sc.pAreas, p_area)
}

func (p_area *PlainArea) destroy(){ }

func (sc *Screen) NewPlainArea(xywh XYWH) *PlainArea{
	var p_area PlainArea 
	sc.initArea(&p_area, xywh)
	return &p_area
}



// --- DrawArea --- //

type DrawArea struct {
	xwindow  		C.xcb_window_t
    cairo_surface	*C.cairo_surface_t
    cairo_context	*C.cairo_t
}

func (d_area *DrawArea) getXWindow() C.xcb_window_t {
	return d_area.xwindow
}

func (d_area *DrawArea) setXWindow(xwindow C.xcb_window_t) {
	d_area.xwindow = xwindow
}

func (d_area *DrawArea) register(sc *Screen){
	sc.dAreas = append(sc.dAreas, d_area)
}

func (d_area *DrawArea) destroy(){
    C.cairo_destroy(d_area.cairo_context)
    C.cairo_surface_destroy(d_area.cairo_surface)
}

func (sc *Screen) NewDrawArea(xywh XYWH) *DrawArea{
	var d_area DrawArea
	sc.initArea(&d_area, xywh)

	visual_is_found := false
	var visual_type *C.xcb_visualtype_t
	
	for depth_itr := C.xcb_screen_allowed_depths_iterator(sc.xscreen); visual_is_found == false && depth_itr.rem != 0; C.xcb_depth_next(&depth_itr) {
		for visual_itr := C.xcb_depth_visuals_iterator(depth_itr.data); visual_is_found == false && visual_itr.rem != 0; C.xcb_visualtype_next(&visual_itr) {
            if (sc.xscreen.root_visual == visual_itr.data.visual_id) {
                visual_type = visual_itr.data
				visual_is_found = true
            }
        }
	}

	if visual_type != nil {
		d_area.cairo_surface = C.cairo_xcb_surface_create(sc.connection, d_area.xwindow, visual_type, C.int(xywh.W), C.int(xywh.H))
		d_area.cairo_context = C.cairo_create(d_area.cairo_surface)
	}

	return &d_area
}

