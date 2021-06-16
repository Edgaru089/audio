package ogg

// #cgo linux darwin LDFLAGS: -lvorbisfile -lvorbisenc -lvorbis
//
// #cgo windows        CFLAGS: -I./extlib/include
// #cgo windows,386   LDFLAGS: -L./extlib/lib-mingw-32 -lvorbisfile -lvorbisenc -lvorbis
// #cgo windows,amd64 LDFLAGS: -L./extlib/lib-mingw-64 -lvorbisfile -lvorbisenc -lvorbis
import "C"
