package main

/*
#cgo CFLAGS: -I${SRCDIR}/lpac/euicc

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
)

func NewLibeuicc(apdu APDU) (*Libeuicc, error) {
	libeuicc := &Libeuicc{
		euiccCtx: (*C.struct_euicc_ctx)(C.malloc(C.sizeof_struct_euicc_ctx)),
	}
	initAPDU(libeuicc.euiccCtx, apdu)
	initHttp(libeuicc.euiccCtx)
	if C.euicc_init(libeuicc.euiccCtx) == CError {
		return nil, ErrEuiccInitFailed
	}
	return libeuicc, nil
}

func (l *Libeuicc) Free() {
	C.euicc_fini(l.euiccCtx)
	C.free(unsafe.Pointer(l.euiccCtx))
}
