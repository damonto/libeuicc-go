package main

import "C"
import "unsafe"

func GoStrings(cStrings **C.char) []string {
	var result []string
	for i := 0; ; i++ {
		cStr := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cStrings)) + uintptr(i)*unsafe.Sizeof(*cStrings)))
		if cStr == nil {
			break
		}
		goStr := C.GoString(cStr)
		result = append(result, goStr)
	}
	return result
}
