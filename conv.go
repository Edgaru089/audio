package audio

// #include <malloc.h>
// #include <string.h>
import "C"
import (
	"reflect"
	"unsafe"
)

// ptrf returns the pointer in *C.float, pointing to the first element of the slice.
//
// If s is empty, this function returns nil.
func ptrf(s []float32) *C.float {
	if len(s) == 0 {
		return nil
	}
	return (*C.float)((unsafe.Pointer)(&s[0]))
}

// c_str returns the *C.char of the string, adding a ending \0.
//
// It is to be freed by calling free.
func c_str(str string) (c *C.char, free func()) {
	str_internal := (*reflect.StringHeader)((unsafe.Pointer)(&str))
	ptr := unsafe.Pointer(C.malloc(C.size_t(len(str) + 1)))

	C.memcpy(ptr, unsafe.Pointer(str_internal.Data), C.size_t(len(str)))

	return (*C.char)(ptr), func() { C.free(ptr) }
}

// c_str_const returns the *C.char address of the string.
//
// The caller should make sure the string is not garbage collected.
//
// the string should have a \0 in itself.
func c_str_const(str string) *C.char {
	str_internal := (*reflect.StringHeader)((unsafe.Pointer)(&str))

	return (*C.char)((unsafe.Pointer)(str_internal.Data))
}
