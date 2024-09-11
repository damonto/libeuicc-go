package libeuicc

/*
#include <stdlib.h>
#include <string.h>

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

func New(driver APDU, customLogger Logger) (*Libeuicc, error) {
	if customLogger != nil {
		logger = customLogger
	}

	euiccCtx := (*C.struct_euicc_ctx)(C.malloc(C.sizeof_struct_euicc_ctx))
	if euiccCtx == nil {
		return nil, ErrNotEnoughMemory
	}
	C.memset(unsafe.Pointer(euiccCtx), 0, C.sizeof_struct_euicc_ctx)

	libeuicc := &Libeuicc{
		ctx: euiccCtx,
	}

	if err := libeuicc.initAPDU(driver); err != nil {
		libeuicc.Close()
		return nil, err
	}
	if err := libeuicc.initHttp(); err != nil {
		libeuicc.Close()
		return nil, err
	}

	if C.euicc_init(libeuicc.ctx) == CError {
		return nil, ErrEuiccInitFailed
	}
	return libeuicc, nil
}

func (e *Libeuicc) Close() {
	if e.ctx != nil {
		C.euicc_fini(e.ctx)
		defer func() {
			C.free(unsafe.Pointer(e.ctx))
			e.ctx = nil
		}()
	}
	if e.ctx.http._interface != nil {
		if e.ctx.http._interface.userdata != nil {
			e.ctx.http._interface.userdata = nil
		}
		C.free(unsafe.Pointer(e.ctx.http._interface))
		e.ctx.http._interface = nil
	}
	if e.ctx.apdu._interface != nil {
		if e.ctx.apdu._interface.userdata != nil {
			e.ctx.apdu._interface.userdata = nil
		}
		C.free(unsafe.Pointer(e.ctx.apdu._interface))
		e.ctx.apdu._interface = nil
	}
}

func (e *Libeuicc) cleanupHttp() {
	C.euicc_http_cleanup(e.ctx)
}
