package flac

// #cgo linux darwin LDFLAGS: -lFLAC -logg
//
// #cgo windows        CFLAGS: -I./extlib/include
// #cgo windows,386   LDFLAGS: -L./extlib/lib-mingw-32 -lFLAC -logg
// #cgo windows,amd64 LDFLAGS: -L./extlib/lib-mingw-64 -lFLAC -logg
import "C"
