package main

import "C"
import (
	"unsafe"
)

func GoStrings(cStrings **C.char) []string {
	var slice []string
	for _, s := range (*[1 << 28]*C.char)(unsafe.Pointer(cStrings))[:2:2] {
		slice = append(slice, C.GoString(s))
	}
	return slice
}
