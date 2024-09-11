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
	euiccCtx   *C.struct_euicc_ctx
	ApduDriver *ApduDriver
}

type ApduDriver struct {
	driver APDU
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
		euiccCtx: euiccCtx,
		ApduDriver: &ApduDriver{
			driver: driver,
		},
	}

	if err := libeuicc.initAPDU(); err != nil {
		libeuicc.Close()
		return nil, err
	}
	if err := libeuicc.initHttp(); err != nil {
		libeuicc.Close()
		return nil, err
	}

	if C.euicc_init(libeuicc.euiccCtx) == CError {
		return nil, ErrEuiccInitFailed
	}
	return libeuicc, nil
}

func (e *Libeuicc) Close() {
	if e.euiccCtx != nil {
		C.euicc_fini(e.euiccCtx)
		defer func() {
			C.free(unsafe.Pointer(e.euiccCtx))
			e.euiccCtx = nil
		}()
	}
	if e.euiccCtx.http._interface != nil {
		C.free(unsafe.Pointer(e.euiccCtx.http._interface))
		e.euiccCtx.http._interface = nil
	}
	if e.euiccCtx.apdu._interface != nil {
		C.free(unsafe.Pointer(e.euiccCtx.apdu._interface))
		e.euiccCtx.apdu._interface = nil
	}
}

func (e *Libeuicc) cleanupHttp() {
	C.euicc_http_cleanup(e.euiccCtx)
}
