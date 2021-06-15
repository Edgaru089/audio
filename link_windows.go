// +build windows

package audio

// #cgo CFLAGS: -I./extlib/include
// #cgo 386   LDFLAGS: -L./extlib/lib-mingw-32 -lopenal32
// #cgo amd64 LDFLAGS: -L./extlib/lib-mingw-64 -lopenal32
import "C"
