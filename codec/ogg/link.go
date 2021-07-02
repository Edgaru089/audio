package ogg

// #cgo linux darwin LDFLAGS: -lvorbisfile -lvorbisenc -lvorbis -logg
//
// #cgo windows        CFLAGS: -I./extlib/include
// #cgo windows,386   LDFLAGS: -L./extlib/lib-mingw-32 -lvorbisfile -lvorbisenc -lvorbis -logg
// #cgo windows,amd64 LDFLAGS: -L./extlib/lib-mingw-64 -lvorbisfile -lvorbisenc -lvorbis -logg
import "C"
