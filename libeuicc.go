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
	ctx *C.struct_euicc_ctx
}

var (
	ErrEuiccInitFailed = errors.New("euicc_init failed")
	ErrNotEnoughMemory = errors.New("not enough memory")
)

func NewLibeuicc(driver APDU, customLogger Logger) (*Libeuicc, error) {
	if customLogger != nil {
		logger = customLogger
	}

	euiccCtx := (*C.struct_euicc_ctx)(C.malloc(C.sizeof_struct_euicc_ctx))
	if euiccCtx == nil {
		return nil, ErrNotEnoughMemory
	}

	libeuicc := &Libeuicc{
		ctx: euiccCtx,
	}
	libeuicc.initAPDU(driver)
	libeuicc.initHttp()

	if C.euicc_init(libeuicc.ctx) == CError {
		return nil, ErrEuiccInitFailed
	}
	return libeuicc, nil
}

func (e *Libeuicc) Free() {
	e.cleanupHttp()
	C.euicc_fini(e.ctx)
	if e.ctx.http._interface != nil {
		C.free(unsafe.Pointer(e.ctx.http._interface))
	}
	if e.ctx.apdu._interface != nil {
		C.free(unsafe.Pointer(e.ctx.apdu._interface))
	}
	C.free(unsafe.Pointer(e.ctx))
}

func (e *Libeuicc) cleanupHttp() {
	C.euicc_http_cleanup(e.ctx)
}
