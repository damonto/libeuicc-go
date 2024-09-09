package libeuicc

/*
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
	ErrEuiccInitFailed = errors.New("euicc_init failed")
	ErrNotEnoughMemory = errors.New("not enough memory")
)

func NewLibeuicc(apdu APDU) (*Libeuicc, error) {
	euiccCtx := (*C.struct_euicc_ctx)(C.malloc(C.sizeof_struct_euicc_ctx))
	if euiccCtx == nil {
		return nil, ErrNotEnoughMemory
	}

	libeuicc := &Libeuicc{
		euiccCtx: euiccCtx,
	}

	libeuicc.initAPDU(apdu)
	libeuicc.initHttp()

	if C.euicc_init(libeuicc.euiccCtx) == CError {
		return nil, ErrEuiccInitFailed
	}
	return libeuicc, nil
}

func (e *Libeuicc) Free() {
	e.cleanupHttp()
	C.euicc_fini(e.euiccCtx)
	if e.euiccCtx.http._interface != nil {
		C.free(unsafe.Pointer(e.euiccCtx.http._interface))
	}
	if e.euiccCtx.apdu._interface != nil {
		C.free(unsafe.Pointer(e.euiccCtx.apdu._interface))
	}
	C.free(unsafe.Pointer(e.euiccCtx))
}

func (e *Libeuicc) cleanupHttp() {
	C.euicc_http_cleanup(e.euiccCtx)
}
