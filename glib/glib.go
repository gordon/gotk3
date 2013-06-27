/*
 * Copyright (c) 2013 Conformal Systems <info@conformal.com>
 *
 * This file originated from: http://opensource.conformal.com/
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

/*
Go bindings for GLib 2.  Supports version 2.36 and later.
*/
package glib

// #cgo pkg-config: glib-2.0 gobject-2.0
// #include <glib.h>
// #include <glib-object.h>
// #include "glib.go.h"
import "C"
import (
	"errors"
	"reflect"
	"runtime"
	"unsafe"
)

var (
	callbackContexts []*CallbackContext
)

/*
 * Type conversions
 */

func gbool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}
func gobool(b C.gboolean) bool {
	if b != 0 {
		return true
	}
	return false
}

/*
 * Unexported vars
 */

var nilPtrErr = errors.New("cgo returned unexpected nil pointer")

/*
 * Constants
 */

type Type int

const _TYPE_FUNDAMENTAL_SHIFT = 2

const (
	TYPE_INVALID Type = iota << _TYPE_FUNDAMENTAL_SHIFT
	TYPE_NONE
	TYPE_INTERFACE
	TYPE_CHAR
	TYPE_UCHAR
	TYPE_BOOLEAN
	TYPE_INT
	TYPE_UINT
	TYPE_LONG
	TYPE_ULONG
	TYPE_INT64
	TYPE_UINT64
	TYPE_ENUM
	TYPE_FLAGS
	TYPE_FLOAT
	TYPE_DOUBLE
	TYPE_STRING
	TYPE_POINTER
	TYPE_BOXED
	TYPE_PARAM
	TYPE_OBJECT
)

/*
 * Events
 */

type CallbackContext struct {
	f      interface{}
	cbi    unsafe.Pointer
	target reflect.Value
	data   reflect.Value
}

type CallbackArg uintptr

func (c *CallbackContext) Target() interface{} {
	return c.target.Interface()
}

func (c *CallbackContext) Data() interface{} {
	return c.data.Interface()
}

func (c *CallbackContext) Arg(n int) CallbackArg {
	return CallbackArg(C.cbinfo_get_arg((*C.cbinfo)(c.cbi), C.int(n)))
}

func (c CallbackArg) String() string {
	return C.GoString((*C.char)(unsafe.Pointer(c)))
}

func (c CallbackArg) Int() int {
	return int(C.int(C.uintptr_t(c)))
}

func (c CallbackArg) UInt() uint {
	return uint(C.uint(C.uintptr_t(c)))
}

//export _go_glib_callback
func _go_glib_callback(cbi *C.cbinfo) {
	ctx := callbackContexts[int(cbi.func_n)]
	rf := reflect.ValueOf(ctx.f)
	t := rf.Type()
	fargs := make([]reflect.Value, t.NumIn())
	if len(fargs) > 0 {
		fargs[0] = reflect.ValueOf(ctx)
	}
	ret := rf.Call(fargs)
	if len(ret) > 0 {
		bret, _ := ret[0].Interface().(bool)
		cbi.ret = gbool(bret)
	}
}

/*
 * GObject
 */

type IObject interface {
	toGObject() *C.GObject
}

type Object struct {
	GObject *C.GObject
}

func (v *Object) Native() *C.GObject {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGObject(p)
}

func (v *Object) toGObject() *C.GObject {
	if v == nil {
		return nil
	}
	return v.Native()
}

func ToGObject(p unsafe.Pointer) *C.GObject {
	return C.toGObject(p)
}

func (v *Object) Ref() {
	C.g_object_ref(C.gpointer(v.GObject))
}

func (v *Object) Unref() {
	C.g_object_unref(C.gpointer(v.GObject))
}

func (v *Object) RefSink() {
	C.g_object_ref_sink(C.gpointer(v.GObject))
}

func (v *Object) IsFloating() bool {
	c := C.g_object_is_floating(C.gpointer(v.GObject))
	return gobool(c)
}

func (v *Object) ForceFloating() {
	C.g_object_force_floating(v.GObject)
}

func (v *Object) StopEmission(s string) {
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))
	C.g_signal_stop_emission_by_name((C.gpointer)(v.GObject),
		(*C.gchar)(cstr))
}

func (v *Object) connectCtx(ctx *CallbackContext, s string, f interface{}) int {
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))
	ctx.cbi = unsafe.Pointer(C._g_signal_connect(unsafe.Pointer(v.GObject),
		(*C.gchar)(cstr), C.int(len(callbackContexts))))
	callbackContexts = append(callbackContexts, ctx)
	return len(callbackContexts) - 1
}

func (v *Object) Connect(s string, f interface{}) int {
	ctx := &CallbackContext{f, nil, reflect.ValueOf(v),
		reflect.ValueOf(nil)}
	return v.connectCtx(ctx, s, f)
}

func (v *Object) ConnectWithData(s string, f interface{}, data interface{}) int {
	ctx := &CallbackContext{f, nil, reflect.ValueOf(v),
		reflect.ValueOf(data)}
	return v.connectCtx(ctx, s, f)
}

// Unlike g_object_set(), this function only sets one name value pair.
// Make multiple calls to set multiple properties.
func (v *Object) Set(name string, value interface{}) {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	if _, ok := value.(Object); ok {
		value = value.(Object).GObject
	}

	var p unsafe.Pointer
	switch value.(type) {
	case bool:
		c := gbool(value.(bool))
		p = unsafe.Pointer(&c)
	case byte:
		c := C.int(value.(byte))
		p = unsafe.Pointer(&c)
	case int:
		c := C.int(value.(int))
		p = unsafe.Pointer(&c)
	case uint:
		c := C.int(value.(uint))
		p = unsafe.Pointer(&c)
	case float32:
		c := C.int(value.(float32))
		p = unsafe.Pointer(&c)
	case float64:
		c := C.int(value.(float64))
		p = unsafe.Pointer(&c)
	case string:
		cstr := C.CString(value.(string))
		defer C.free(unsafe.Pointer(cstr))
		p = unsafe.Pointer(cstr)
	default:
		if pv, ok := value.(unsafe.Pointer); ok {
			p = pv
		} else {
			val := reflect.ValueOf(value)
			c := C.int(val.Int())
			p = unsafe.Pointer(&c)
		}
	}
	// Can't call g_object_set() as it uses a variable arg list, use a
	// wrapper instead
	C._g_object_set_one(C.gpointer(v.GObject), (*C.gchar)(cstr), p)
}

/*
 * GObject Signals
 */

func (v *Object) Emit(s string) {
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))
	C._g_signal_emit_by_name_one((C.gpointer)(v.GObject), (*C.gchar)(cstr))
}

func (v *Object) HandlerBlock(callID int) {
	id := C.cbinfo_get_id((*C.cbinfo)(callbackContexts[callID].cbi))
	C.g_signal_handler_block((C.gpointer)(v.GObject), id)
}

func (v *Object) HandlerUnblock(callID int) {
	id := C.cbinfo_get_id((*C.cbinfo)(callbackContexts[callID].cbi))
	C.g_signal_handler_unblock((C.gpointer)(v.GObject), id)
}

func (v *Object) HandlerDisconnect(callID int) {
	id := C.cbinfo_get_id((*C.cbinfo)(callbackContexts[callID].cbi))
	C.g_signal_handler_disconnect((C.gpointer)(v.GObject), id)
}

/*
 * GInitiallyUnowned
 */

type InitiallyUnowned struct {
	*Object
}

/*
 * GValue
 */

// Don't allocate Values on the stack or heap manually as they may not
// be properly unset when going out of scope. Instead, use ValueAlloc(),
// which will set the runtime finalizer to unset the Value.
type Value struct {
	GValue C.GValue
}

func (v *Value) Native() *C.GValue {
	return &v.GValue
}

func ValueAlloc() (*Value, error) {
	c := C._g_value_alloc()
	if c == nil {
		return nil, nilPtrErr
	}
	v := &Value{*c}
	runtime.SetFinalizer(v, (*Value).unset)
	return v, nil
}

func ValueInit(t Type) (*Value, error) {
	c := C._g_value_init(C.GType(t))
	if c == nil {
		return nil, nilPtrErr
	}
	v := &Value{*c}
	runtime.SetFinalizer(v, (*Value).unset)
	return v, nil
}

func (v *Value) unset() {
	C.g_value_unset(v.Native())
}

func (v *Value) GetType() Type {
	c := C.g_value_get_gtype(v.Native())
	return Type(c)
}

// Converts a native Go type to the comparable GValue.
func GValue(v interface{}) (gvalue *Value, err error) {
	switch v.(type) {
	case int:
		val, err := ValueInit(TYPE_INT)
		if err != nil {
			return nil, err
		}
		val.SetInt(v.(int))
		return val, nil
	case string:
		val, err := ValueInit(TYPE_STRING)
		if err != nil {
			return nil, err
		}
		val.SetString(v.(string))
		return val, nil
	default:
		return nil, errors.New("Type not implemented")
	}
	return nil, nil
}

// Converts a GValue to comparable Go type
func (v *Value) GoValue() (interface{}, error) {
	switch v.GetType() {
	case TYPE_INT:
		c := C.g_value_get_int(v.Native())
		return int(c), nil
	case TYPE_STRING:
		c := C.g_value_get_string(v.Native())
		return C.GoString((*C.char)(c)), nil
	default:
		return nil, errors.New("Type conversion not supported")
	}
}

func (v *Value) SetInt(val int) {
	C.g_value_set_int(v.Native(), C.gint(val))
}

func (v *Value) SetString(val string) {
	cstr := C.CString(val)
	defer C.free(unsafe.Pointer(cstr))
	C.g_value_set_string(v.Native(), (*C.gchar)(cstr))
}

func (v *Value) GetString() (string, error) {
	c := C.g_value_get_string(v.Native())
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}