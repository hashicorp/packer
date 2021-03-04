// +build !safe
// +build !appengine
// +build go1.9

// Copyright (c) 2012-2020 Ugorji Nwoke. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

// minimum of go 1.9 is needed, as that is the minimum for all features and linked functions we need
// - typedmemclr was introduced in go 1.8
// - mapassign_fastXXX was introduced in go 1.9
// etc

package codec

import (
	"reflect"
	_ "runtime" // needed so that gccgo works with go linkname(s)
	"sync/atomic"
	"time"
	"unsafe"
)

// This file has unsafe variants of some helper functions.
// MARKER: See helper_unsafe.go for the usage documentation.

// For reflect.Value code, we decided to do the following:
//    - if we know the kind, we can elide conditional checks for
//      - SetXXX (Int, Uint, String, Bool, etc)
//      - SetLen
//
// We can also optimize
//      - IsNil

// MARKER: Some functions here will not be hit during code coverage runs due to optimizations, e.g.
//   - rvCopySlice:      decode calls it if rvGrowSlice didn't set the new slice into the pointer to the orig slice.
//                       however, helper_unsafe sets it, so there's no need to call rvCopySlice later
//   - rvSlice:          same as above
//   - rvGetArray4Bytes: only called within kArray for []byte, but that is now handled
//                       within the fast-path directly

const safeMode = false

// helperUnsafeDirectAssignMapEntry says that we should not copy the pointer in the map
// to another value during mapRange/iteration and mapGet calls, but directly assign it.
//
// The only callers of mapRange/iteration is encode.
// Here, we just walk through the values and encode them
//
// The only caller of mapGet is decode.
// Here, it does a Get if the underlying value is a pointer, and decodes into that.
//
// For both users, we are very careful NOT to modify or keep the pointers around.
// Consequently, it is ok for take advantage of the performance that the map is not modified
// during an iteration and we can just "peek" at the internal value" in the map and use it.
const helperUnsafeDirectAssignMapEntry = true

// MARKER: keep in sync with GO_ROOT/src/reflect/value.go
const (
	unsafeFlagStickyRO = 1 << 5
	unsafeFlagEmbedRO  = 1 << 6
	unsafeFlagIndir    = 1 << 7
	unsafeFlagAddr     = 1 << 8
	// unsafeFlagKindMask = (1 << 5) - 1 // 5 bits for 27 kinds (up to 31)
	// unsafeTypeKindDirectIface = 1 << 5
)

type unsafeString struct {
	Data unsafe.Pointer
	Len  int
}

type unsafeSlice struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

type unsafeIntf struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

type unsafeReflectValue struct {
	unsafeIntf
	flag uintptr
}

// keep in sync with stdlib runtime/type.go
type unsafeRuntimeType struct {
	size uintptr
	// ... many other fields here
}

// unsafeZeroScalarXXX below is used in rvZeroAddrPrimK as the backing storage
// for primitives that we want to decode into.
//
// MARKER: slices are not supported, as they are mutable, and thus an element can be being
// decoded when we are using this space for a slice that should hold it.
//
// Cap is 16 as the maximum size is a complex128 (or string on 64-bit machines).

const unsafeZeroScalarArrCap = 16

var (
	unsafeZeroScalarArr [2][unsafeZeroScalarArrCap]byte

	unsafeZeroScalar0Addr = &unsafeZeroScalarArr[0]
	unsafeZeroScalar1Addr = &unsafeZeroScalarArr[1]
)

// unsafeZeroAddr and unsafeZeroSlice points to a read-only block of memory
// used for setting a zero value for most types or creating a read-only
// zero value for a given type.

var (
	unsafeZeroAddr  = unsafe.Pointer(&unsafeZeroArr[0])
	unsafeZeroSlice = unsafeSlice{unsafeZeroAddr, 0, 0}
)

// stringView returns a view of the []byte as a string.
// In unsafe mode, it doesn't incur allocation and copying caused by conversion.
// In regular safe mode, it is an allocation and copy.
func stringView(v []byte) string {
	return *(*string)(unsafe.Pointer(&v))
}

// bytesView returns a view of the string as a []byte.
// In unsafe mode, it doesn't incur allocation and copying caused by conversion.
// In regular safe mode, it is an allocation and copy.
func bytesView(v string) (b []byte) {
	sx := (*unsafeString)(unsafe.Pointer(&v))
	bx := (*unsafeSlice)(unsafe.Pointer(&b))
	bx.Data, bx.Len, bx.Cap = sx.Data, sx.Len, sx.Len
	return
}

func byteSliceSameData(v1 []byte, v2 []byte) bool {
	return (*unsafeSlice)(unsafe.Pointer(&v1)).Data == (*unsafeSlice)(unsafe.Pointer(&v2)).Data
}

// isNil says whether the value v is nil.
// This applies to references like map/ptr/unsafepointer/chan/func,
// and non-reference values like interface/slice.
func isNil(v interface{}) (rv reflect.Value, isnil bool) {
	var ui = (*unsafeIntf)(unsafe.Pointer(&v))
	if ui.ptr == nil {
		isnil = true
		return
	}
	rv = reflect.ValueOf(v) // reflect.ValueOf is currently not inline'able - so call it directly
	tk := rv.Kind()
	isnil = (tk == reflect.Interface || tk == reflect.Slice) && *(*unsafe.Pointer)(ui.ptr) == nil
	return
}

// return the pointer for a reference (map/chan/func/pointer/unsafe.Pointer).
// true references (map, func, chan, ptr - NOT slice) may be double-referenced? as flagIndir
//
// Assumes that v is a reference (map/func/chan/ptr/func)
func rvRefPtr(v *unsafeReflectValue) unsafe.Pointer {
	if v.flag&unsafeFlagIndir != 0 {
		return *(*unsafe.Pointer)(v.ptr)
	}
	return v.ptr
}

// func rvAddr(rv reflect.Value) uintptr {
// 	return uintptr((*unsafeReflectValue)(unsafe.Pointer(&rv)).ptr)
// }

func eq4i(i0, i1 interface{}) bool {
	v0 := (*unsafeIntf)(unsafe.Pointer(&i0))
	v1 := (*unsafeIntf)(unsafe.Pointer(&i1))
	return v0.typ == v1.typ && v0.ptr == v1.ptr
}

// func rv4i(i interface{}) reflect.Value {
// 	// Unfortunately, we cannot get the "kind" of the interface directly here.
// 	// We need the 'rtype', whose structure changes in different go versions.
// 	// Finally, it's not clear that there is benefit to reimplementing it,
// 	// as the "escapes(i)" is not clearly expensive since we want i to exist on the heap.
//
// 	return reflect.ValueOf(i)
// }

// func rv4iK(i interface{}, kind byte, isref bool) (v reflect.Value) {
// 	// Main advantage here is that it is inlined, nothing escapes to heap, i is never nil
// 	uv := (*unsafeReflectValue)(unsafe.Pointer(&v))
// 	uv.unsafeIntf = *(*unsafeIntf)(unsafe.Pointer(&i))
// 	uv.flag = uintptr(kind)
// 	if !isref {
// 		uv.flag |= unsafeFlagIndir
// 	}
// 	return
// }

func rv4iptr(i interface{}) (v reflect.Value) {
	// Main advantage here is that it is inlined, nothing escapes to heap, i is never nil
	uv := (*unsafeReflectValue)(unsafe.Pointer(&v))
	uv.unsafeIntf = *(*unsafeIntf)(unsafe.Pointer(&i))
	uv.flag = uintptr(rkindPtr)
	return
}

func rv4istr(i interface{}) (v reflect.Value) {
	// Main advantage here is that it is inlined, nothing escapes to heap, i is never nil
	uv := (*unsafeReflectValue)(unsafe.Pointer(&v))
	uv.unsafeIntf = *(*unsafeIntf)(unsafe.Pointer(&i))
	uv.flag = uintptr(rkindString) | unsafeFlagIndir
	return
}

// func rv2ptr(urv *unsafeReflectValue) unsafe.Pointer {
// 	if refBitset.isset(byte(urv.flag&unsafeFlagKindMask)) && urv.flag&unsafeFlagIndir != 0 {
// 		return *(*unsafe.Pointer)(urv.ptr)
// 	}
// 	return urv.ptr
// }

// func rv2i(rv reflect.Value) interface{} {
// 	// We tap into implememtation details from
// 	// the source go stdlib reflect/value.go, and trims the implementation.
// 	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
// 	return *(*interface{})(unsafe.Pointer(&unsafeIntf{typ: urv.typ, word: rv2ptr(urv)}))
// }

func rv2i(rv reflect.Value) (i interface{}) {
	// We tap into implememtation details from
	// the source go stdlib reflect/value.go, and trims the implementation.
	//
	// e.g.
	// - a map/ptr is a reference,        thus flagIndir is not set on it
	// - an int/slice is not a reference, thus flagIndir is set on it

	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	if refBitset.isset(byte(rv.Kind())) && urv.flag&unsafeFlagIndir != 0 {
		urv.ptr = *(*unsafe.Pointer)(urv.ptr)
	}
	return *(*interface{})(unsafe.Pointer(&urv.unsafeIntf))
}

func rvIsNil(rv reflect.Value) bool {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	if urv.flag&unsafeFlagIndir != 0 {
		return *(*unsafe.Pointer)(urv.ptr) == nil
	}
	return urv.ptr == nil
}

func rvSetSliceLen(rv reflect.Value, length int) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	(*unsafeString)(urv.ptr).Len = length
}

func rvZeroAddrK(t reflect.Type, k reflect.Kind) (rv reflect.Value) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	urv.typ = ((*unsafeIntf)(unsafe.Pointer(&t))).ptr
	urv.flag = uintptr(k) | unsafeFlagIndir | unsafeFlagAddr
	urv.ptr = unsafe_New(urv.typ)
	return
}

// rvZeroAddrTransientK is used for getting a *transient* value to be decoded into,
// which will right away be used for something else.
//
// For this, we optimize and use a scratch space if the kind is a number, bool or string.
//
// We use this for situations:
// - we need to decode into a primitive for setting into an interface.
// - decode into a primitive for setting a map value
// - decode into a primitive for setting a map key iff the map value is also a primitive
//
// For all these, the decoding can be thought of as a one-off value which will not
// interfer with other values being decoded.
//
// Because of the situation with map keys and map values being primitives, we have 2 variants:
// Transient and Transient2 (used for map keys if map value is primitive also)
func rvZeroAddrTransientAnyK(t reflect.Type, k reflect.Kind, addr *[unsafeZeroScalarArrCap]byte) (rv reflect.Value) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	urv.typ = ((*unsafeIntf)(unsafe.Pointer(&t))).ptr
	urv.flag = uintptr(k) | unsafeFlagIndir | unsafeFlagAddr
	// if k is number, bool or string, use scratch space
	if scalarBitset.isset(byte(k)) {
		*addr = [unsafeZeroScalarArrCap]byte{}
		urv.ptr = unsafe.Pointer(addr)
	} else {
		urv.ptr = unsafe_New(urv.typ)
	}
	return
}

func rvZeroAddrTransientK(t reflect.Type, k reflect.Kind) reflect.Value {
	return rvZeroAddrTransientAnyK(t, k, unsafeZeroScalar0Addr)
}

func rvZeroAddrTransient2K(t reflect.Type, k reflect.Kind) reflect.Value {
	return rvZeroAddrTransientAnyK(t, k, unsafeZeroScalar1Addr)
}

func rvZeroK(t reflect.Type, k reflect.Kind) (rv reflect.Value) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	urv.typ = ((*unsafeIntf)(unsafe.Pointer(&t))).ptr
	if refBitset.isset(byte(k)) {
		urv.flag = uintptr(k)
	} else if (k == reflect.Struct || k == reflect.Array) && rtsize2(urv.typ) > uintptr(len(unsafeZeroArr)) {
		urv.flag = uintptr(k) | unsafeFlagIndir | unsafeFlagAddr
		urv.ptr = unsafe_New(urv.typ)
	} else {
		urv.flag = uintptr(k) | unsafeFlagIndir
		urv.ptr = unsafeZeroAddr
	}
	return
}

// rvConvert will convert a value to a different type directly,
// ensuring that they still point to the same underlying value.
func rvConvert(v reflect.Value, t reflect.Type) reflect.Value {
	uv := (*unsafeReflectValue)(unsafe.Pointer(&v))
	uv.typ = ((*unsafeIntf)(unsafe.Pointer(&t))).ptr
	return v
}

// rvAddressableReadonly returns an addressable reflect.Value.
//
// Use it within encode calls, when you just want to "read" the underlying ptr
// without modifying the value.
//
// Note that it cannot be used for r/w use, as those non-addressable values
// may have been stored in read-only memory, and trying to write the pointer
// may cause a segfault.
func rvAddressableReadonly(v reflect.Value) reflect.Value {
	// hack to make an addressable value out of a non-addressable one.
	// Assume folks calling it are passing a value that can be addressable, but isn't.
	// This assumes that the flagIndir is already set on it.
	// so we just set the flagAddr bit on the flag (and do not set the flagIndir).

	uv := (*unsafeReflectValue)(unsafe.Pointer(&v))
	uv.flag = uv.flag | unsafeFlagAddr // | unsafeFlagIndir

	// callers of this only use it in read-only mode
	// if uv.ptr == nil || uv.ptr == unsafeZeroAddr {
	// 	uv.ptr = unsafe_New(uv.typ)
	// }

	return v

	// rv := rvZeroAddrK(v.Type(), v.Kind())
	// rvSetDirect(rv, v)
	// return rv.Addr()
}

func rtsize2(rt unsafe.Pointer) uintptr {
	return ((*unsafeRuntimeType)(rt)).size
}

func rt2id(rt reflect.Type) uintptr {
	return uintptr(((*unsafeIntf)(unsafe.Pointer(&rt))).ptr)
}

func i2rtid(i interface{}) uintptr {
	return uintptr(((*unsafeIntf)(unsafe.Pointer(&i))).typ)
}

// --------------------------

func unsafeCmpZero(ptr unsafe.Pointer, size int) bool {
	var s1 = unsafeString{ptr, size}
	var s2 = unsafeString{unsafeZeroAddr, size}
	if size > len(unsafeZeroArr) {
		arr := make([]byte, size)
		s2.Data = unsafe.Pointer(&arr[0])
	}
	return *(*string)(unsafe.Pointer(&s1)) == *(*string)(unsafe.Pointer(&s2)) // memcmp
}

func isEmptyValue(v reflect.Value, tinfos *TypeInfos, recursive bool) bool {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&v))
	if urv.flag == 0 {
		return true
	}
	if recursive {
		return isEmptyValueFallbackRecur(urv, v, tinfos)
	}
	// t := rvPtrToType(urv.typ)
	// // it is empty if it is a zero value OR it points to a zero value
	// if urv.flag&unsafeFlagIndir == 0 { // this is a pointer
	// 	if urv.ptr == nil {
	// 		return true
	// 	}
	// 	return unsafeCmpZero(*(*unsafe.Pointer)(urv.ptr), int(rtsize(t.Elem())))
	// }
	// return unsafeCmpZero(urv.ptr, int(rtsize(t)))
	// return unsafeCmpZero(urv.ptr, int(rtsize(rvPtrToType(urv.typ))))

	// return unsafeCmpZero(urv.ptr, int(rtsize(rvType(v))))
	return unsafeCmpZero(urv.ptr, int(rtsize2(urv.typ)))

}

func isEmptyValueFallbackRecur(urv *unsafeReflectValue, v reflect.Value, tinfos *TypeInfos) bool {
	const recursive = true

	switch v.Kind() {
	case reflect.Invalid:
		return true
	case reflect.String:
		return (*unsafeString)(urv.ptr).Len == 0
	case reflect.Slice:
		return (*unsafeSlice)(urv.ptr).Len == 0
	case reflect.Bool:
		return !*(*bool)(urv.ptr)
	case reflect.Int:
		return *(*int)(urv.ptr) == 0
	case reflect.Int8:
		return *(*int8)(urv.ptr) == 0
	case reflect.Int16:
		return *(*int16)(urv.ptr) == 0
	case reflect.Int32:
		return *(*int32)(urv.ptr) == 0
	case reflect.Int64:
		return *(*int64)(urv.ptr) == 0
	case reflect.Uint:
		return *(*uint)(urv.ptr) == 0
	case reflect.Uint8:
		return *(*uint8)(urv.ptr) == 0
	case reflect.Uint16:
		return *(*uint16)(urv.ptr) == 0
	case reflect.Uint32:
		return *(*uint32)(urv.ptr) == 0
	case reflect.Uint64:
		return *(*uint64)(urv.ptr) == 0
	case reflect.Uintptr:
		return *(*uintptr)(urv.ptr) == 0
	case reflect.Float32:
		return *(*float32)(urv.ptr) == 0
	case reflect.Float64:
		return *(*float64)(urv.ptr) == 0
	case reflect.Complex64:
		return unsafeCmpZero(urv.ptr, 8)
	case reflect.Complex128:
		return unsafeCmpZero(urv.ptr, 16)
	case reflect.Struct:
		// return isEmptyStruct(v, tinfos, recursive)
		if tinfos == nil {
			tinfos = defTypeInfos
		}
		ti := tinfos.find(uintptr(urv.typ))
		if ti == nil {
			ti = tinfos.load(rvType(v))
		}
		return unsafeCmpZero(urv.ptr, int(ti.size))
	case reflect.Interface, reflect.Ptr:
		// isnil := urv.ptr == nil // (not sufficient, as a pointer value encodes the type)
		isnil := urv.ptr == nil || *(*unsafe.Pointer)(urv.ptr) == nil
		if recursive && !isnil {
			return isEmptyValue(v.Elem(), tinfos, recursive)
		}
		return isnil
	case reflect.UnsafePointer:
		return urv.ptr == nil || *(*unsafe.Pointer)(urv.ptr) == nil
	case reflect.Chan:
		return urv.ptr == nil || len_chan(rvRefPtr(urv)) == 0
	case reflect.Map:
		return urv.ptr == nil || len_map(rvRefPtr(urv)) == 0
	case reflect.Array:
		return v.Len() == 0
	}
	return false
}

// --------------------------

// atomicXXX is expected to be 2 words (for symmetry with atomic.Value)
//
// Note that we do not atomically load/store length and data pointer separately,
// as this could lead to some races. Instead, we atomically load/store cappedSlice.
//
// Note: with atomic.(Load|Store)Pointer, we MUST work with an unsafe.Pointer directly.

// ----------------------
type atomicTypeInfoSlice struct {
	v unsafe.Pointer // *[]rtid2ti
	// _ uint64         // padding (atomicXXX expected to be 2 words)
}

func (x *atomicTypeInfoSlice) load() (s []rtid2ti) {
	x2 := atomic.LoadPointer(&x.v)
	if x2 != nil {
		s = *(*[]rtid2ti)(x2)
	}
	return
}

func (x *atomicTypeInfoSlice) store(p []rtid2ti) {
	atomic.StorePointer(&x.v, unsafe.Pointer(&p))
}

// MARKER: in safe mode, atomicXXX are atomic.Value, which contains an interface{}.
// This is 2 words.
// consider padding atomicXXX here with a uintptr, so they fit into 2 words also.

// --------------------------
type atomicRtidFnSlice struct {
	v unsafe.Pointer // *[]codecRtidFn
}

func (x *atomicRtidFnSlice) load() (s []codecRtidFn) {
	x2 := atomic.LoadPointer(&x.v)
	if x2 != nil {
		s = *(*[]codecRtidFn)(x2)
	}
	return
}

func (x *atomicRtidFnSlice) store(p []codecRtidFn) {
	atomic.StorePointer(&x.v, unsafe.Pointer(&p))
}

// --------------------------
type atomicClsErr struct {
	v unsafe.Pointer // *clsErr
}

func (x *atomicClsErr) load() (e clsErr) {
	x2 := (*clsErr)(atomic.LoadPointer(&x.v))
	if x2 != nil {
		e = *x2
	}
	return
}

func (x *atomicClsErr) store(p clsErr) {
	atomic.StorePointer(&x.v, unsafe.Pointer(&p))
}

// --------------------------

// to create a reflect.Value for each member field of fauxUnion,
// we first create a global fauxUnion, and create reflect.Value
// for them all.
// This way, we have the flags and type in the reflect.Value.
// Then, when a reflect.Value is called, we just copy it,
// update the ptr to the fauxUnion's, and return it.

type unsafeDecNakedWrapper struct {
	fauxUnion
	ru, ri, rf, rl, rs, rb, rt reflect.Value // mapping to the primitives above
}

func (n *unsafeDecNakedWrapper) init() {
	n.ru = rv4iptr(&n.u).Elem()
	n.ri = rv4iptr(&n.i).Elem()
	n.rf = rv4iptr(&n.f).Elem()
	n.rl = rv4iptr(&n.l).Elem()
	n.rs = rv4iptr(&n.s).Elem()
	n.rt = rv4iptr(&n.t).Elem()
	n.rb = rv4iptr(&n.b).Elem()
	// n.rr[] = reflect.ValueOf(&n.)
}

var defUnsafeDecNakedWrapper unsafeDecNakedWrapper

func init() {
	defUnsafeDecNakedWrapper.init()
}

func (n *fauxUnion) ru() (v reflect.Value) {
	v = defUnsafeDecNakedWrapper.ru
	((*unsafeReflectValue)(unsafe.Pointer(&v))).ptr = unsafe.Pointer(&n.u)
	return
}
func (n *fauxUnion) ri() (v reflect.Value) {
	v = defUnsafeDecNakedWrapper.ri
	((*unsafeReflectValue)(unsafe.Pointer(&v))).ptr = unsafe.Pointer(&n.i)
	return
}
func (n *fauxUnion) rf() (v reflect.Value) {
	v = defUnsafeDecNakedWrapper.rf
	((*unsafeReflectValue)(unsafe.Pointer(&v))).ptr = unsafe.Pointer(&n.f)
	return
}
func (n *fauxUnion) rl() (v reflect.Value) {
	v = defUnsafeDecNakedWrapper.rl
	((*unsafeReflectValue)(unsafe.Pointer(&v))).ptr = unsafe.Pointer(&n.l)
	return
}
func (n *fauxUnion) rs() (v reflect.Value) {
	v = defUnsafeDecNakedWrapper.rs
	((*unsafeReflectValue)(unsafe.Pointer(&v))).ptr = unsafe.Pointer(&n.s)
	return
}
func (n *fauxUnion) rt() (v reflect.Value) {
	v = defUnsafeDecNakedWrapper.rt
	((*unsafeReflectValue)(unsafe.Pointer(&v))).ptr = unsafe.Pointer(&n.t)
	return
}
func (n *fauxUnion) rb() (v reflect.Value) {
	v = defUnsafeDecNakedWrapper.rb
	((*unsafeReflectValue)(unsafe.Pointer(&v))).ptr = unsafe.Pointer(&n.b)
	return
}

// --------------------------
func rvSetBytes(rv reflect.Value, v []byte) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	*(*[]byte)(urv.ptr) = v
}

func rvSetString(rv reflect.Value, v string) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	*(*string)(urv.ptr) = v
}

func rvSetBool(rv reflect.Value, v bool) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	*(*bool)(urv.ptr) = v
}

func rvSetTime(rv reflect.Value, v time.Time) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	*(*time.Time)(urv.ptr) = v
}

func rvSetFloat32(rv reflect.Value, v float32) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	*(*float32)(urv.ptr) = v
}

func rvSetFloat64(rv reflect.Value, v float64) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	*(*float64)(urv.ptr) = v
}

func rvSetComplex64(rv reflect.Value, v complex64) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	*(*complex64)(urv.ptr) = v
}

func rvSetComplex128(rv reflect.Value, v complex128) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	*(*complex128)(urv.ptr) = v
}

func rvSetInt(rv reflect.Value, v int) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	*(*int)(urv.ptr) = v
}

func rvSetInt8(rv reflect.Value, v int8) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	*(*int8)(urv.ptr) = v
}

func rvSetInt16(rv reflect.Value, v int16) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	*(*int16)(urv.ptr) = v
}

func rvSetInt32(rv reflect.Value, v int32) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	*(*int32)(urv.ptr) = v
}

func rvSetInt64(rv reflect.Value, v int64) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	*(*int64)(urv.ptr) = v
}

func rvSetUint(rv reflect.Value, v uint) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	*(*uint)(urv.ptr) = v
}

func rvSetUintptr(rv reflect.Value, v uintptr) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	*(*uintptr)(urv.ptr) = v
}

func rvSetUint8(rv reflect.Value, v uint8) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	*(*uint8)(urv.ptr) = v
}

func rvSetUint16(rv reflect.Value, v uint16) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	*(*uint16)(urv.ptr) = v
}

func rvSetUint32(rv reflect.Value, v uint32) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	*(*uint32)(urv.ptr) = v
}

func rvSetUint64(rv reflect.Value, v uint64) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	*(*uint64)(urv.ptr) = v
}

// ----------------

// rvSetDirect is rv.Set for all kinds except reflect.Interface.
//
// Callers MUST not pass an interface value in, as it may result in an unexpected segfaults.
func rvSetDirect(rv reflect.Value, v reflect.Value) {
	// MARKER: rv.Set for interface may need to do a separate allocation if a scalar value.
	// The book-keeping is onerous, so we just do the simple ones where a memmove is sufficient.
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	uv := (*unsafeReflectValue)(unsafe.Pointer(&v))
	if uv.flag&unsafeFlagIndir == 0 {
		*(*unsafe.Pointer)(urv.ptr) = uv.ptr
	} else if uv.ptr == unsafeZeroAddr {
		if urv.ptr != unsafeZeroAddr {
			typedmemclr(urv.typ, urv.ptr)
		}
	} else {
		typedmemmove(urv.typ, urv.ptr, uv.ptr)
	}
}

func rvSetDirectZero(rv reflect.Value) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	if urv.ptr != unsafeZeroAddr {
		typedmemclr(urv.typ, urv.ptr)
	}
}

// // rvSlice returns a slice of the slice of lenth
// func rvSlice(rv reflect.Value, length int) (v reflect.Value) {
// 	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
// 	uv := (*unsafeReflectValue)(unsafe.Pointer(&v))
// 	*uv = *urv
// 	var x []unsafe.Pointer
// 	uv.ptr = unsafe.Pointer(&x)
// 	*(*unsafeSlice)(uv.ptr) = *(*unsafeSlice)(urv.ptr)
// 	(*unsafeSlice)(uv.ptr).Len = length
// 	return
// }

// rvMakeSlice updates the slice to point to a new array.
// It copies data from old slice to new slice.
// It returns set=true iff it updates it, else it just returns a new slice pointing to a newly made array.
func rvMakeSlice(rv reflect.Value, ti *typeInfo, xlen, xcap int) (_ reflect.Value, set bool) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	ux := (*unsafeSlice)(urv.ptr)
	t := ((*unsafeIntf)(unsafe.Pointer(&ti.elem))).ptr
	s := unsafeSlice{unsafe_NewArray(t, xcap), xlen, xcap}
	if ux.Len > 0 {
		typedslicecopy(t, s, *ux)
	}
	*ux = s
	return rv, true
}

// rcGrowSlice updates the slice to point to a new array with the cap incremented, and len set to the new cap value.
// It copies data from old slice to new slice.
// It returns set=true iff it updates it, else it just returns a new slice pointing to a newly made array.
func rvGrowSlice(rv reflect.Value, ti *typeInfo, xcap, incr int) (v reflect.Value, newcap int, set bool) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	ux := (*unsafeSlice)(urv.ptr)
	t := ((*unsafeIntf)(unsafe.Pointer(&ti.elem))).ptr
	*ux = growslice(t, *ux, xcap+incr)
	ux.Len = ux.Cap
	return rv, ux.Cap, true
}

// rvSlice returns a sub-slice of the slice given new lenth,
// without modifying passed in value.
// It is typically called when we know that SetLen(...) cannot be done.
func rvSlice(rv reflect.Value, length int) reflect.Value {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	var x []struct{}
	ux := (*unsafeSlice)(unsafe.Pointer(&x))
	*ux = *(*unsafeSlice)(urv.ptr)
	ux.Len = length
	urv.ptr = unsafe.Pointer(ux)
	return rv
}

// ------------

func rvSliceIndex(rv reflect.Value, i int, ti *typeInfo) (v reflect.Value) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	uv := (*unsafeReflectValue)(unsafe.Pointer(&v))
	uv.ptr = unsafe.Pointer(uintptr(((*unsafeSlice)(urv.ptr)).Data) + uintptr(int(ti.elemsize)*i))
	uv.typ = ((*unsafeIntf)(unsafe.Pointer(&ti.elem))).ptr
	uv.flag = uintptr(ti.elemkind) | unsafeFlagIndir | unsafeFlagAddr
	return
}

func rvSliceZeroCap(t reflect.Type) (v reflect.Value) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&v))
	urv.typ = ((*unsafeIntf)(unsafe.Pointer(&t))).ptr
	urv.flag = uintptr(reflect.Slice) | unsafeFlagIndir
	urv.ptr = unsafe.Pointer(&unsafeZeroSlice)
	return
}

func rvLenSlice(rv reflect.Value) int {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return (*unsafeSlice)(urv.ptr).Len
}

func rvCapSlice(rv reflect.Value) int {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return (*unsafeSlice)(urv.ptr).Cap
}

func rvArrayIndex(rv reflect.Value, i int, ti *typeInfo) (v reflect.Value) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	uv := (*unsafeReflectValue)(unsafe.Pointer(&v))
	uv.ptr = unsafe.Pointer(uintptr(urv.ptr) + uintptr(int(ti.elemsize)*i))
	uv.typ = ((*unsafeIntf)(unsafe.Pointer(&ti.elem))).ptr
	uv.flag = uintptr(ti.elemkind) | unsafeFlagIndir | unsafeFlagAddr
	return
}

// if scratch is nil, then return a writable view (assuming canAddr=true)
func rvGetArrayBytes(rv reflect.Value, scratch []byte) (bs []byte) {
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	bx := (*unsafeSlice)(unsafe.Pointer(&bs))
	bx.Data = urv.ptr
	bx.Len = rv.Len()
	bx.Cap = bx.Len
	return
}

func rvGetArray4Slice(rv reflect.Value) (v reflect.Value) {
	// It is possible that this slice is based off an array with a larger
	// len that we want (where array len == slice cap).
	// However, it is ok to create an array type that is a subset of the full
	// e.g. full slice is based off a *[16]byte, but we can create a *[4]byte
	// off of it. That is ok.
	//
	// Consequently, we use rvLenSlice, not rvCapSlice.

	t := reflectArrayOf(rvLenSlice(rv), rvType(rv).Elem())
	// v = rvZeroAddrK(t, reflect.Array)

	uv := (*unsafeReflectValue)(unsafe.Pointer(&v))
	uv.flag = uintptr(reflect.Array) | unsafeFlagIndir | unsafeFlagAddr
	uv.typ = ((*unsafeIntf)(unsafe.Pointer(&t))).ptr

	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	uv.ptr = *(*unsafe.Pointer)(urv.ptr) // slice rv has a ptr to the slice.

	return
}

func rvGetSlice4Array(rv reflect.Value, v interface{}) {
	// v is a pointer to a slice to be populated
	uv := (*unsafeIntf)(unsafe.Pointer(&v))
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))

	s := (*unsafeSlice)(uv.ptr)
	s.Data = urv.ptr
	s.Len = rv.Len()
	s.Cap = s.Len
	return
}

func rvCopySlice(dest, src reflect.Value) {
	t := rvType(dest).Elem()
	urv := (*unsafeReflectValue)(unsafe.Pointer(&dest))
	destPtr := urv.ptr
	urv = (*unsafeReflectValue)(unsafe.Pointer(&src))
	typedslicecopy((*unsafeIntf)(unsafe.Pointer(&t)).ptr,
		*(*unsafeSlice)(destPtr), *(*unsafeSlice)(urv.ptr))
}

// ------------

func rvGetBool(rv reflect.Value) bool {
	v := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return *(*bool)(v.ptr)
}

func rvGetBytes(rv reflect.Value) []byte {
	v := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return *(*[]byte)(v.ptr)
}

func rvGetTime(rv reflect.Value) time.Time {
	v := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return *(*time.Time)(v.ptr)
}

func rvGetString(rv reflect.Value) string {
	v := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return *(*string)(v.ptr)
}

func rvGetFloat64(rv reflect.Value) float64 {
	v := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return *(*float64)(v.ptr)
}

func rvGetFloat32(rv reflect.Value) float32 {
	v := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return *(*float32)(v.ptr)
}

func rvGetComplex64(rv reflect.Value) complex64 {
	v := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return *(*complex64)(v.ptr)
}

func rvGetComplex128(rv reflect.Value) complex128 {
	v := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return *(*complex128)(v.ptr)
}

func rvGetInt(rv reflect.Value) int {
	v := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return *(*int)(v.ptr)
}

func rvGetInt8(rv reflect.Value) int8 {
	v := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return *(*int8)(v.ptr)
}

func rvGetInt16(rv reflect.Value) int16 {
	v := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return *(*int16)(v.ptr)
}

func rvGetInt32(rv reflect.Value) int32 {
	v := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return *(*int32)(v.ptr)
}

func rvGetInt64(rv reflect.Value) int64 {
	v := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return *(*int64)(v.ptr)
}

func rvGetUint(rv reflect.Value) uint {
	v := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return *(*uint)(v.ptr)
}

func rvGetUint8(rv reflect.Value) uint8 {
	v := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return *(*uint8)(v.ptr)
}

func rvGetUint16(rv reflect.Value) uint16 {
	v := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return *(*uint16)(v.ptr)
}

func rvGetUint32(rv reflect.Value) uint32 {
	v := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return *(*uint32)(v.ptr)
}

func rvGetUint64(rv reflect.Value) uint64 {
	v := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return *(*uint64)(v.ptr)
}

func rvGetUintptr(rv reflect.Value) uintptr {
	v := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	return *(*uintptr)(v.ptr)
}

func rvLenMap(rv reflect.Value) int {
	// maplen is not inlined, because as of go1.16beta, go:linkname's are not inlined.
	// thus, faster to call rv.Len() directly.
	//
	// MARKER: review after https://github.com/golang/go/issues/20019 fixed.

	// return rv.Len()

	return len_map(rvRefPtr((*unsafeReflectValue)(unsafe.Pointer(&rv))))
}

// Note: it is hard to find len(...) of an array type,
// as that is a field in the arrayType representing the array,
// and hard to introspect.
//
// func rvLenArray(rv reflect.Value) int {	return rv.Len() }

// ------------ map range and map indexing ----------

// regular calls to map via reflection: MapKeys, MapIndex, MapRange/MapIter etc
// will always allocate for each map key or value.
//
// It is more performant to provide a value that the map entry is set into,
// and that elides the allocation.

// go 1.4+ has runtime/hashmap.go or runtime/map.go which has a
// hIter struct with the first 2 values being key and value
// of the current iteration.
//
// This *hIter is passed to mapiterinit, mapiternext, mapiterkey, mapiterelem.
// We bypass the reflect wrapper functions and just use the *hIter directly.
//
// Though *hIter has many fields, we only care about the first 2.
//
// We directly embed this in unsafeMapIter below
//
// hiter is typically about 12 words, but we just fill up unsafeMapIter to 32 words,
// so it fills multiple cache lines and can give some extra space to accomodate small growth.

type unsafeMapIter struct {
	mtyp, mptr unsafe.Pointer
	k, v       reflect.Value
	kisref     bool
	visref     bool
	mapvalues  bool
	done       bool
	started    bool
	_          [3]byte // padding
	it         struct {
		key   unsafe.Pointer
		value unsafe.Pointer
		_     [20]uintptr // padding for other fields (to make up 32 words for enclosing struct)
	}
}

func (t *unsafeMapIter) Next() (r bool) {
	if t == nil || t.done {
		return
	}
	if t.started {
		mapiternext((unsafe.Pointer)(&t.it))
	} else {
		t.started = true
	}

	t.done = t.it.key == nil
	if t.done {
		return
	}

	if helperUnsafeDirectAssignMapEntry || t.kisref {
		(*unsafeReflectValue)(unsafe.Pointer(&t.k)).ptr = t.it.key
	} else {
		k := (*unsafeReflectValue)(unsafe.Pointer(&t.k))
		typedmemmove(k.typ, k.ptr, t.it.key)
	}

	if t.mapvalues {
		if helperUnsafeDirectAssignMapEntry || t.visref {
			(*unsafeReflectValue)(unsafe.Pointer(&t.v)).ptr = t.it.value
		} else {
			v := (*unsafeReflectValue)(unsafe.Pointer(&t.v))
			typedmemmove(v.typ, v.ptr, t.it.value)
		}
	}

	return true
}

func (t *unsafeMapIter) Key() (r reflect.Value) {
	return t.k
}

func (t *unsafeMapIter) Value() (r reflect.Value) {
	return t.v
}

func (t *unsafeMapIter) Done() {}

type mapIter struct {
	unsafeMapIter
}

func mapRange(t *mapIter, m, k, v reflect.Value, mapvalues bool) {
	if rvIsNil(m) {
		t.done = true
		return
	}
	t.done = false
	t.started = false
	t.mapvalues = mapvalues

	// var urv *unsafeReflectValue

	urv := (*unsafeReflectValue)(unsafe.Pointer(&m))
	t.mtyp = urv.typ
	t.mptr = rvRefPtr(urv)

	// t.it = (*unsafeMapHashIter)(reflect_mapiterinit(t.mtyp, t.mptr))
	mapiterinit(t.mtyp, t.mptr, unsafe.Pointer(&t.it))

	t.k = k
	t.kisref = refBitset.isset(byte(k.Kind()))

	if mapvalues {
		t.v = v
		t.visref = refBitset.isset(byte(v.Kind()))
	} else {
		t.v = reflect.Value{}
	}
}

// // unsafeMapSet does equivalent of: p = p2
// func unsafeMapSet(ptyp, p, p2 unsafe.Pointer, isref bool) {
// 	if isref {
// 		*(*unsafe.Pointer)(p) = *(*unsafe.Pointer)(p2) // p2
// 	} else {
// 		typedmemmove(ptyp, p, p2) // *(*unsafe.Pointer)(p2)) // p2)
// 	}
// }

// unsafeMapKVPtr returns the pointer if flagIndir, else it returns a pointer to the pointer.
// It is needed as maps always keep a reference to the underlying value.
func unsafeMapKVPtr(urv *unsafeReflectValue) unsafe.Pointer {
	if urv.flag&unsafeFlagIndir == 0 {
		return unsafe.Pointer(&urv.ptr)
	}
	return urv.ptr
}

func mapGet(m, k, v reflect.Value, keyFastKind mapKeyFastKind, valIsIndirect, valIsRef bool) (_ reflect.Value) {
	var urv = (*unsafeReflectValue)(unsafe.Pointer(&k))
	var kptr = unsafeMapKVPtr(urv)
	urv = (*unsafeReflectValue)(unsafe.Pointer(&m))
	mptr := rvRefPtr(urv)

	var vvptr unsafe.Pointer
	var ok bool

	// Note that mapaccess2_fastXXX functions do not check if the value needs to be copied.
	// if they do, we should dereference the pointer and return that

	switch keyFastKind {
	case mapKeyFastKind32, mapKeyFastKind32ptr:
		vvptr, ok = mapaccess2_fast32(urv.typ, mptr, *(*uint32)(kptr))
	case mapKeyFastKind64, mapKeyFastKind64ptr:
		vvptr, ok = mapaccess2_fast64(urv.typ, mptr, *(*uint64)(kptr))
	case mapKeyFastKindStr:
		vvptr, ok = mapaccess2_faststr(urv.typ, mptr, *(*string)(kptr))
	default:
		vvptr, ok = mapaccess2(urv.typ, mptr, kptr)
	}

	if !ok {
		return
	}

	urv = (*unsafeReflectValue)(unsafe.Pointer(&v))

	if keyFastKind != 0 && valIsIndirect {
		urv.ptr = *(*unsafe.Pointer)(vvptr)
	} else if helperUnsafeDirectAssignMapEntry || valIsRef {
		urv.ptr = vvptr
	} else {
		typedmemmove(urv.typ, urv.ptr, vvptr)
	}

	return v
}

func mapSet(m, k, v reflect.Value, keyFastKind mapKeyFastKind, valIsIndirect, valIsRef bool) {
	var urv = (*unsafeReflectValue)(unsafe.Pointer(&k))
	var kptr = unsafeMapKVPtr(urv)
	urv = (*unsafeReflectValue)(unsafe.Pointer(&v))
	var vtyp = urv.typ
	var vptr = unsafeMapKVPtr(urv)

	urv = (*unsafeReflectValue)(unsafe.Pointer(&m))
	mptr := rvRefPtr(urv)

	var vvptr unsafe.Pointer

	// mapassign_fastXXX don't take indirect into account.
	// It was hard to infer what makes it work all the time.
	// Sometimes, we got vvptr == nil when we dereferenced vvptr (if valIsIndirect).
	// Consequently, only use fastXXX functions if !valIsIndirect

	const alwaysUseGenericMapassign = false

	if alwaysUseGenericMapassign || valIsIndirect {
		vvptr = mapassign(urv.typ, mptr, kptr)
		typedmemmove(vtyp, vvptr, vptr)
		// reflect_mapassign(urv.typ, mptr, kptr, vptr)
		return
	}

	switch keyFastKind {
	case mapKeyFastKind32:
		vvptr = mapassign_fast32(urv.typ, mptr, *(*uint32)(kptr))
	case mapKeyFastKind32ptr:
		vvptr = mapassign_fast32ptr(urv.typ, mptr, *(*unsafe.Pointer)(kptr))
	case mapKeyFastKind64:
		vvptr = mapassign_fast64(urv.typ, mptr, *(*uint64)(kptr))
	case mapKeyFastKind64ptr:
		vvptr = mapassign_fast64ptr(urv.typ, mptr, *(*unsafe.Pointer)(kptr))
	case mapKeyFastKindStr:
		vvptr = mapassign_faststr(urv.typ, mptr, *(*string)(kptr))
	default:
		vvptr = mapassign(urv.typ, mptr, kptr)
	}

	// if keyFastKind != 0 && valIsIndirect {
	// 	vvptr = *(*unsafe.Pointer)(vvptr)
	// }

	typedmemmove(vtyp, vvptr, vptr)
}

// func mapDelete(m, k reflect.Value) {
// 	var urv = (*unsafeReflectValue)(unsafe.Pointer(&k))
// 	var kptr = unsafeMapKVPtr(urv)
// 	urv = (*unsafeReflectValue)(unsafe.Pointer(&m))
// 	mapdelete(urv.typ, rv2ptr(urv), kptr)
// }

// return an addressable reflect value that can be used in mapRange and mapGet operations.
//
// all calls to mapGet or mapRange will call here to get an addressable reflect.Value.
func mapAddrLoopvarRV(t reflect.Type, k reflect.Kind) (rv reflect.Value) {
	// return rvZeroAddrK(t, k)
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	urv.flag = uintptr(k) | unsafeFlagIndir | unsafeFlagAddr
	urv.typ = ((*unsafeIntf)(unsafe.Pointer(&t))).ptr
	// since we always set the ptr when helperUnsafeDirectAssignMapEntry=true,
	// we should only allocate if it is not true
	if !helperUnsafeDirectAssignMapEntry {
		urv.ptr = unsafe_New(urv.typ)
	}
	return
}

// ---------- ENCODER optimized ---------------

func (e *Encoder) jsondriver() *jsonEncDriver {
	return (*jsonEncDriver)((*unsafeIntf)(unsafe.Pointer(&e.e)).ptr)
}

// ---------- DECODER optimized ---------------

func (d *Decoder) checkBreak() bool {
	// MARKER: jsonDecDriver.CheckBreak() costs over 80, and this isn't inlined.
	// Consequently, there's no benefit in incurring the cost of this
	// wrapping function checkBreak.
	//
	// It is faster to just call the interface method directly.

	// if d.js {
	// 	return d.jsondriver().CheckBreak()
	// }
	// if d.cbor {
	// 	return d.cbordriver().CheckBreak()
	// }
	return d.d.CheckBreak()
}

func (d *Decoder) jsondriver() *jsonDecDriver {
	return (*jsonDecDriver)((*unsafeIntf)(unsafe.Pointer(&d.d)).ptr)
}

// ---------- structFieldInfo optimized ---------------

func (n *structFieldInfoPathNode) rvField(v reflect.Value) (rv reflect.Value) {
	// we already know this is exported, and maybe embedded (based on what si says)
	uv := (*unsafeReflectValue)(unsafe.Pointer(&v))
	urv := (*unsafeReflectValue)(unsafe.Pointer(&rv))
	// clear flagEmbedRO if necessary, and inherit permission bits from v
	urv.flag = uv.flag&(unsafeFlagStickyRO|unsafeFlagIndir|unsafeFlagAddr) | uintptr(n.kind)
	urv.typ = ((*unsafeIntf)(unsafe.Pointer(&n.typ))).ptr
	urv.ptr = unsafe.Pointer(uintptr(uv.ptr) + uintptr(n.offset))
	return
}

// runtime chan and map are designed such that the first field is the count.
// len builtin uses this to get the length of a chan/map easily.
// leverage this knowledge, since maplen and chanlen functions from runtime package
// are go:linkname'd here, and thus not inlined as of go1.16beta

func len_map_chan(m unsafe.Pointer) int {
	if m == nil {
		return 0
	}
	return *((*int)(m))
}

func len_map(m unsafe.Pointer) int {
	// return maplen(m)
	return len_map_chan(m)
}
func len_chan(m unsafe.Pointer) int {
	// return chanlen(m)
	return len_map_chan(m)
}

// ---------- go linknames (LINKED to runtime/reflect) ---------------

// MARKER: always check that these linknames match subsequent versions of go
//
// Note that as of Jan 2021 (go 1.16 release), go:linkname(s) are not inlined
// outside of the standard library use (e.g. within sync, reflect, etc).
// If these link'ed functions were normally inlined, calling them here would
// not necessarily give a performance boost, due to function overhead.
//
// However, it seems most of these functions are not inlined anyway,
// as only maplen, chanlen and mapaccess are small enough to get inlined.
//
//   We checked this by going into $GOROOT/src/runtime and running:
//   $ go build -tags notfastpath -gcflags "-m=2"
//
// Also, we link to the functions in reflect where possible, as opposed to those in runtime.
// They are guaranteed to be safer for our use, even when they are just trampoline functions.

//go:linkname mapiterinit runtime.mapiterinit
//go:noescape
func mapiterinit(typ unsafe.Pointer, m unsafe.Pointer, it unsafe.Pointer)

//go:linkname mapiternext runtime.mapiternext
//go:noescape
func mapiternext(it unsafe.Pointer) (key unsafe.Pointer)

//go:linkname mapaccess2 runtime.mapaccess2
//go:noescape
func mapaccess2(typ unsafe.Pointer, m unsafe.Pointer, key unsafe.Pointer) (val unsafe.Pointer, ok bool)

//go:linkname mapaccess2_fast32 runtime.mapaccess2_fast32
//go:noescape
func mapaccess2_fast32(typ unsafe.Pointer, m unsafe.Pointer, key uint32) (val unsafe.Pointer, ok bool)

//go:linkname mapaccess2_fast64 runtime.mapaccess2_fast64
//go:noescape
func mapaccess2_fast64(typ unsafe.Pointer, m unsafe.Pointer, key uint64) (val unsafe.Pointer, ok bool)

//go:linkname mapaccess2_faststr runtime.mapaccess2_faststr
//go:noescape
func mapaccess2_faststr(typ unsafe.Pointer, m unsafe.Pointer, key string) (val unsafe.Pointer, ok bool)

//go:linkname mapassign_fast32 runtime.mapassign_fast32
//go:noescape
func mapassign_fast32(typ unsafe.Pointer, m unsafe.Pointer, key uint32) unsafe.Pointer

//go:linkname mapassign_fast32ptr runtime.mapassign_fast32ptr
//go:noescape
func mapassign_fast32ptr(typ unsafe.Pointer, m unsafe.Pointer, key unsafe.Pointer) unsafe.Pointer

//go:linkname mapassign_fast64 runtime.mapassign_fast64
//go:noescape
func mapassign_fast64(typ unsafe.Pointer, m unsafe.Pointer, key uint64) unsafe.Pointer

//go:linkname mapassign_fast64ptr runtime.mapassign_fast64ptr
//go:noescape
func mapassign_fast64ptr(typ unsafe.Pointer, m unsafe.Pointer, key unsafe.Pointer) unsafe.Pointer

//go:linkname mapassign_faststr runtime.mapassign_faststr
//go:noescape
func mapassign_faststr(typ unsafe.Pointer, m unsafe.Pointer, s string) unsafe.Pointer

//go:linkname mapassign runtime.mapassign
//go:noescape
func mapassign(typ unsafe.Pointer, m unsafe.Pointer, key unsafe.Pointer) unsafe.Pointer

//go:linkname mapdelete reflect.mapdelete
//go:noescape
func mapdelete(typ unsafe.Pointer, m unsafe.Pointer, key unsafe.Pointer)

//go:linkname unsafe_New reflect.unsafe_New
//go:noescape
func unsafe_New(typ unsafe.Pointer) unsafe.Pointer

//go:linkname unsafe_NewArray reflect.unsafe_NewArray
//go:noescape
func unsafe_NewArray(typ unsafe.Pointer, cap int) unsafe.Pointer

//go:linkname typedslicecopy reflect.typedslicecopy
//go:noescape
func typedslicecopy(elemType unsafe.Pointer, dst, src unsafeSlice) int

//go:linkname typedmemmove reflect.typedmemmove
//go:noescape
func typedmemmove(typ unsafe.Pointer, dst, src unsafe.Pointer)

//go:linkname typedmemclr reflect.typedmemclr
//go:noescape
func typedmemclr(typ unsafe.Pointer, dst unsafe.Pointer)

//go:linkname growslice runtime.growslice
//go:noescape
func growslice(typ unsafe.Pointer, old unsafeSlice, cap int) unsafeSlice

/*

//go:linkname maplen reflect.maplen
//go:noescape
func maplen(typ unsafe.Pointer) int

//go:linkname chanlen reflect.chanlen
//go:noescape
func chanlen(typ unsafe.Pointer) int

//go:linkname reflect_mapiterinit reflect.mapiterinit
//go:noescape
func reflect_mapiterinit(typ unsafe.Pointer, m unsafe.Pointer) (it unsafe.Pointer)

//go:linkname reflect_mapaccess reflect.mapaccess
//go:noescape
func reflect_mapaccess(typ unsafe.Pointer, m unsafe.Pointer, key unsafe.Pointer) (val unsafe.Pointer)

//go:linkname reflect_mapassign reflect.mapassign
//go:noescape
func reflect_mapassign(typ unsafe.Pointer, m unsafe.Pointer, key, val unsafe.Pointer)

//go:linkname memhash runtime.memhash
//go:noescape
func memhash(p unsafe.Pointer, seed, length uintptr) uintptr

// ---------- others ---------------

func hashShortString(b []byte) uintptr {
	return memhash(unsafe.Pointer(&b[0]), 0, uintptr(len(b)))
}

func rtsize(rt reflect.Type) uintptr {
	return ((*unsafeRuntimeType)(((*unsafeIntf)(unsafe.Pointer(&rt))).ptr)).size
}

// var _ = runtime.MemProfileRate

*/
