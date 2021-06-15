package audio

// #cgo linux darwin LDFLAGS: -lopenal
//
// #cgo windows        CFLAGS: -I./extlib/include
// #cgo windows,386   LDFLAGS: -L./extlib/lib-mingw-32 -lopenal32
// #cgo windows,amd64 LDFLAGS: -L./extlib/lib-mingw-64 -lopenal32
import "C"
