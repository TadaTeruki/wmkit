package wmkit

/*
#cgo pkg-config: xcb cairo
#include <xcb/xcb.h>
#include <cairo.h>
#include <cairo-xcb.h>
#include <stdlib.h>


void make_window_attributes_value_list(uint32_t* values, size_t size){
	values = (uint32_t*)malloc(sizeof(uint32_t)*size);
}

void set_window_attributes_value_list(uint32_t* values, size_t ad, uint32_t value){
	values[ad] = value;
}

void free_window_attributes_value_list(uint32_t* values){
	free(values);
}


*/
import "C"

import(
	"unsafe"
	cairo "github.com/ungerik/go-cairo"
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

	values := [2]C.uint32_t{ C.uint32_t(1), C.XCB_EVENT_MASK_EXPOSURE | C.XCB_EVENT_MASK_BUTTON_PRESS | C.XCB_EVENT_MASK_BUTTON_RELEASE }

	C.xcb_create_window (sc.connection,
		C.XCB_COPY_FROM_PARENT,
	 	xwindow,
		sc.xscreen.root,
		C.short(xywh.X),  C.short(xywh.Y),
		C.ushort(xywh.W), C.ushort(xywh.H),
		0,
		C.XCB_WINDOW_CLASS_INPUT_OUTPUT,
		sc.xscreen.root_visual,
		C.XCB_CW_OVERRIDE_REDIRECT | C.XCB_CW_EVENT_MASK ,
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
		d_area.cairo_surface = C.cairo_xcb_surface_create(sc.connection, d_area.xwindow, visual_type, C.int(xywh.W), C.int(xywh.H))
		d_area.cairo_context = C.cairo_create(d_area.cairo_surface)
	}

	return &d_area
}

func (d_area *DrawArea) getCairoSurface() *cairo.Surface {
	surface :=
		cairo.NewSurfaceFromC(
			cairo.Cairo_surface(unsafe.Pointer(d_area.cairo_surface)),
			cairo.Cairo_context(unsafe.Pointer(d_area.cairo_context)),
		)
	return surface
}

func (d_area *DrawArea) Draw(f func(*cairo.Surface)){
	f(d_area.getCairoSurface())
}