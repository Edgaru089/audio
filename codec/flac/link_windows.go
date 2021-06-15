// +build windows

package flac

// #cgo CFLAGS: -I./extlib/include
// #cgo 386   LDFLAGS: -L./extlib/lib-mingw-32 -lFLAC -logg
// #cgo amd64 LDFLAGS: -L./extlib/lib-mingw-64 -lFLAC -logg
import "C"
