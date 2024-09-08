package main

/*
#cgo CFLAGS: -I${SRCDIR}/lpac/euicc -I ${SRCDIR}/lpac/cjson
#cgo linux,386 LDFLAGS: -L${SRCDIR}/libs/linux-x86_64 -leuicc -lcjson-static
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/libs/linux-x86_64 -leuicc -lcjson-static

#include <stdlib.h>

#include "euicc.h"
*/
import "C"
import (
	"errors"
	"unsafe"
)

type Libeuicc struct {
	euiccCtx *C.struct_euicc_ctx
}

var (
	ErrEuiccInitFailed      = errors.New("euicc_init failed")
	ErrLibeuiccInvalid      = errors.New("libeuicc path is invalid")
	ErrLibeuiccDoesNotExist = errors.New("libeuicc path does not exist")
	ErrNotEnoughMemory      = errors.New("not enough memory")
)

func NewLibeuicc(apdu APDU) (*Libeuicc, error) {
	euiccCtx := (*C.struct_euicc_ctx)(C.malloc(C.sizeof_struct_euicc_ctx))
	if euiccCtx == nil {
		return nil, ErrNotEnoughMemory
	}

	libeuicc := &Libeuicc{
		euiccCtx: euiccCtx,
	}

	initAPDU(libeuicc.euiccCtx, apdu)
	initHttp(libeuicc.euiccCtx)

	if C.euicc_init(libeuicc.euiccCtx) == CError {
		return nil, ErrEuiccInitFailed
	}
	return libeuicc, nil
}

func (e *Libeuicc) Free() {
	e.cleanupHttp()
	C.euicc_fini(e.euiccCtx)
	C.free(unsafe.Pointer(e.euiccCtx))
}

func (e *Libeuicc) cleanupHttp() {
	C.euicc_http_cleanup(e.euiccCtx)
}
