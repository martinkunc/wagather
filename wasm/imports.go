package wasm

// Based on https://github.com/go-wasm-adapter/go-wasm

/*
#include <stdlib.h>

extern void debug(void *context, int32_t a);
extern void wexit(void *context, int32_t a);
extern void wwrite(void *context, int32_t a);
extern void nanotime(void *context, int32_t a);
extern void nanotime1(void *context, int32_t a);
extern void walltime(void *context, int32_t a);
extern void walltime1(void *context, int32_t a);
extern void scheduleCallback(void *context, int32_t a);
extern void clearScheduledCallback(void *context, int32_t a);
extern void getRandomData(void *context, int32_t a);
extern void stringVal(void *context, int32_t a);
extern void valueInstanceOf(void *context, int32_t a);
extern void valueGet(void *context, int32_t a);
extern void valueSet(void *context, int32_t a);
extern void valueIndex(void *context, int32_t a);
extern void valueSetIndex(void *context, int32_t a);
extern void valueCall(void *context, int32_t a);
extern void valueInvoke(void *context, int32_t a);
extern void valueNew(void *context, int32_t a);
extern void valueLength(void *context, int32_t a);
extern void valuePrepareString(void *context, int32_t a);
extern void valueLoadString(void *context, int32_t a);
extern void valueDelete(void *context, int32_t a);
extern void scheduleTimeoutEvent(void *context, int32_t a);
extern void clearTimeoutEvent(void *context, int32_t a);
extern void copyBytesToGo (void *context, int32_t a);
extern void copyBytesToJS (void *context, int32_t a);
extern void resetMemoryDataView (void *context, int32_t a);
extern void finalizeRef (void *context, int32_t a);
*/
import "C"
import (
	"crypto/rand"
	"fmt"
	"log"
	"reflect"
	"syscall"
	"time"
	"unsafe"

	"github.com/wasmerio/go-ext-wasm/wasmer"
)

//export debug
func debug(_ unsafe.Pointer, sp int32) {
	log.Println(sp)
}

//export wexit
func wexit(ctx unsafe.Pointer, sp int32) {
	b := getBridge(ctx)
	b.exitCode = int(b.getUint32(sp + 8))
	if b.cancF != nil {
		log.Printf("exit code %d\n", b.exitCode)
		b.cancF()
	}
}

//export wwrite
func wwrite(ctx unsafe.Pointer, sp int32) {
	b := getBridge(ctx)
	fd := int(b.getInt64(sp + 8))
	p := int(b.getInt64(sp + 16))
	l := int(b.getInt32(sp + 24))
	_, err := syscall.Write(fd, b.mem()[p:p+l])
	if err != nil {
		panic(fmt.Errorf("wasm-write: %v", err))
	}
}

//export nanotime
func nanotime(ctx unsafe.Pointer, sp int32) {
	n := time.Now().UnixNano()
	getBridge(ctx).setInt64(sp+8, n)
}

//export nanotime1
func nanotime1(ctx unsafe.Pointer, sp int32) {
	n := time.Now().UnixNano()
	getBridge(ctx).setInt64(sp+8, n)
}

//export walltime
func walltime(ctx unsafe.Pointer, sp int32) {
	b := getBridge(ctx)
	t := time.Now().UnixNano()
	nanos := t % int64(time.Second)
	b.setInt64(sp+8, t/int64(time.Second))
	b.setInt32(sp+16, int32(nanos))

}

//export walltime1
func walltime1(ctx unsafe.Pointer, sp int32) {
	b := getBridge(ctx)
	t := time.Now().UnixNano()
	nanos := t % int64(time.Second)
	b.setInt64(sp+8, t/int64(time.Second))
	b.setInt32(sp+16, int32(nanos))
}

//export scheduleCallback
func scheduleCallback(_ unsafe.Pointer, _ int32) {
	panic("schedule callback")
}

//export clearScheduledCallback
func clearScheduledCallback(_ unsafe.Pointer, _ int32) {
	panic("clear scheduled callback")
}

//export getRandomData
func getRandomData(ctx unsafe.Pointer, sp int32) {
	s := getBridge(ctx).loadSlice(sp + 8)
	_, err := rand.Read(s)
	if err != nil {
		panic("failed: getRandomData")
	}
}

//export stringVal
func stringVal(ctx unsafe.Pointer, sp int32) {
	b := getBridge(ctx)
	str := b.loadString(sp + 8)
	b.storeValue(sp+24, str)
}

//export valueInstanceOf
func valueInstanceOf(ctx unsafe.Pointer, sp int32) {
	panic("valueInstanceOf")
	// b := getBridge(ctx)
	// str := b.loadString(sp + 16)
	// val := b.loadValue(sp + 8)
	// sp = b.getSP()
	// obj, ok := val.(*object)
	// if !ok {
	// 	b.storeValue(sp+32, val)
	// 	return
	// }

	// res, ok := obj.props[str]
	// if !ok {
	// 	panic(fmt.Sprintln("missing property", str, val))
	// }
	// b.storeValue(sp+32, res)
}

//export valueGet
func valueGet(ctx unsafe.Pointer, sp int32) {
	b := getBridge(ctx)
	str := b.loadString(sp + 16)
	val := b.loadValue(sp + 8)
	sp = b.getSP()
	obj, ok := val.(*object)
	if !ok {
		b.storeValue(sp+32, val)
		return
	}

	res, ok := obj.props[str]
	if !ok {
		panic(fmt.Sprintln("missing property", str, val))
	}
	b.storeValue(sp+32, res)
}

//export valueSet
func valueSet(ctx unsafe.Pointer, sp int32) {
	b := getBridge(ctx)
	val := b.loadValue(sp + 8)
	obj := val.(*object)
	prop := b.loadString(sp + 16)
	propVal := b.loadValue(sp + 32)
	obj.props[prop] = propVal
}

//export valueIndex
func valueIndex(ctx unsafe.Pointer, sp int32) {
	b := getBridge(ctx)
	l := b.loadValue(sp + 8)
	i := b.getInt64(sp + 16)
	rv := reflect.ValueOf(l)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	iv := rv.Index(int(i))
	b.storeValue(sp+24, iv.Interface())
}

//export valueSetIndex
func valueSetIndex(_ unsafe.Pointer, _ int32) {
	panic("valueSetIndex")
}

//export valueCall
func valueCall(ctx unsafe.Pointer, sp int32) {
	b := getBridge(ctx)
	v := b.loadValue(sp + 8)
	str := b.loadString(sp + 16)
	args := b.loadSliceOfValues(sp + 32)
	f, ok := v.(*object).props[str].(Func)
	if !ok {
		panic(fmt.Sprintf("valueCall: prop not found in %v, %s", v.(*object).name, str))
	}
	sp = b.getSP()
	res, err := f(args)
	if err != nil {
		b.storeValue(sp+56, err.Error())
		b.setUint8(sp+64, 0)
		return
	}

	b.storeValue(sp+56, res)
	b.setUint8(sp+64, 1)
}

//export valueInvoke
func valueInvoke(ctx unsafe.Pointer, sp int32) {
	b := getBridge(ctx)
	val := *(b.loadValue(sp + 8).(*Func))
	args := b.loadSliceOfValues(sp + 16)
	res, err := val(args)
	sp = b.getSP()
	if err != nil {
		b.storeValue(sp+40, err)
		b.setUint8(sp+48, 0)
		return
	}

	b.storeValue(sp+40, res)
	b.setUint8(sp+48, 1)
}

//export valueNew
func valueNew(ctx unsafe.Pointer, sp int32) {
	b := getBridge(ctx)
	val := b.loadValue(sp + 8)
	args := b.loadSliceOfValues(sp + 16)
	res := val.(*object).new(args)
	sp = b.getSP()
	b.storeValue(sp+40, res)
	b.setUint8(sp+48, 1)
}

//export valueLength
func valueLength(ctx unsafe.Pointer, sp int32) {
	b := getBridge(ctx)
	val := b.loadValue(sp + 8)
	rv := reflect.ValueOf(val)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	var l int
	switch {
	case rv.Kind() == reflect.Slice:
		l = rv.Len()
	case rv.Type() == reflect.TypeOf(array{}):
		l = len(val.(*array).buf)
	default:
		panic(fmt.Sprintf("valueLength on %T", val))
	}

	b.setInt64(sp+16, int64(l))
}

//export valuePrepareString
func valuePrepareString(ctx unsafe.Pointer, sp int32) {
	b := getBridge(ctx)
	val := b.loadValue(sp + 8)
	var str string
	if val != nil {
		str = fmt.Sprint(val)
	}

	b.storeValue(sp+16, str)
	b.setInt64(sp+24, int64(len(str)))
}

//export valueLoadString
func valueLoadString(ctx unsafe.Pointer, sp int32) {
	b := getBridge(ctx)
	str := b.loadValue(sp + 8).(string)
	sl := b.loadSlice(sp + 16)
	copy(sl, str)
}

//export valueDelete
func valueDelete(ctx unsafe.Pointer, sp int32) {
	// TODO
}

//export scheduleTimeoutEvent
func scheduleTimeoutEvent(_ unsafe.Pointer, _ int32) {
	panic("scheduleTimeoutEvent")
}

//export clearTimeoutEvent
func clearTimeoutEvent(_ unsafe.Pointer, _ int32) {
	panic("clearTimeoutEvent")
}

//export copyBytesToJS
func copyBytesToJS(ctx unsafe.Pointer, sp int32) {
	b := getBridge(ctx)
	dst, ok := b.loadValue(sp + 8).(*array)
	if !ok {
		b.setUint8(sp+48, 0)
		return
	}
	src := b.loadSlice(sp + 16)
	n := copy(dst.buf, src[:len(dst.buf)])
	b.setInt64(sp+40, int64(n))
	b.setUint8(sp+48, 1)
}

//export copyBytesToGo
func copyBytesToGo(ctx unsafe.Pointer, sp int32) {
	b := getBridge(ctx)
	dst := b.loadSlice(sp + 8)
	src, ok := b.loadValue(sp + 32).(*array)
	if !ok {
		b.setUint8(sp+48, 0)
		return
	}
	n := copy(dst, src.buf[:len(dst)])
	b.setInt64(sp+40, int64(n))
	b.setUint8(sp+48, 1)
}

//export resetMemoryDataView
func resetMemoryDataView(ctx unsafe.Pointer, sp int32) {
	b := getBridge(ctx)
	b.resetMemoryDataViewBridge()
}

//export finalizeRef
func finalizeRef(ctx unsafe.Pointer, sp int32) {
	// ?? This doens't seem to have ref counting
}

// addImports adds go Bridge imports in "go" namespace.
func (b *Bridge) addImports(imps *wasmer.Imports) error {
	imps = imps.Namespace("go")
	var is = []struct {
		name string
		imp  interface{}
		cgo  unsafe.Pointer
	}{
		{"debug", debug, C.debug},
		{"runtime.wasmExit", wexit, C.wexit},
		{"runtime.wasmWrite", wwrite, C.wwrite},
		{"runtime.nanotime", nanotime, C.nanotime},
		{"runtime.nanotime1", nanotime1, C.nanotime1},
		{"runtime.walltime", walltime, C.walltime},
		{"runtime.walltime1", walltime1, C.walltime1},
		{"runtime.scheduleCallback", scheduleCallback, C.scheduleCallback},
		{"runtime.clearScheduledCallback", clearScheduledCallback, C.clearScheduledCallback},
		{"runtime.getRandomData", getRandomData, C.getRandomData},
		{"runtime.scheduleTimeoutEvent", scheduleTimeoutEvent, C.scheduleTimeoutEvent},
		{"runtime.clearTimeoutEvent", clearTimeoutEvent, C.clearTimeoutEvent},
		{"runtime.resetMemoryDataView", resetMemoryDataView, C.resetMemoryDataView},
		{"syscall/js.stringVal", stringVal, C.stringVal},
		{"syscall/js.valueGet", valueGet, C.valueGet},
		{"syscall/js.valueInstanceOf", valueInstanceOf, C.valueInstanceOf},
		{"syscall/js.valueSet", valueSet, C.valueSet},
		{"syscall/js.valueIndex", valueIndex, C.valueIndex},
		{"syscall/js.valueSetIndex", valueSetIndex, C.valueSetIndex},
		{"syscall/js.valueCall", valueCall, C.valueCall},
		{"syscall/js.valueInvoke", valueInvoke, C.valueInvoke},
		{"syscall/js.valueNew", valueNew, C.valueNew},
		{"syscall/js.valueLength", valueLength, C.valueLength},
		{"syscall/js.valuePrepareString", valuePrepareString, C.valuePrepareString},
		{"syscall/js.valueLoadString", valueLoadString, C.valueLoadString},
		{"syscall/js.valueDelete", valueDelete, C.valueDelete},
		{"syscall/js.copyBytesToGo", copyBytesToGo, C.copyBytesToGo},
		{"syscall/js.copyBytesToJS", copyBytesToJS, C.copyBytesToJS},
		{"syscall/js.finalizeRef", finalizeRef, C.finalizeRef},
	}

	var err error
	for _, imp := range is {
		imps, err = imps.Append(imp.name, imp.imp, imp.cgo)
		if err != nil {
			return err
		}
	}

	return nil
}
