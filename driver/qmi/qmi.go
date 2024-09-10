package driver

/*
#cgo pkg-config: glib-2.0 qmi-glib

#include <stdint.h>

#include "qmi.h"
*/
import "C"
import (
	"errors"
	"unsafe"

	"github.com/damonto/libeuicc-go"
)

type qmi struct {
	uimSlot int
	device  string
}

func NewQMI(device string, uimSlot int) libeuicc.APDU {
	return &qmi{
		device:  device,
		uimSlot: uimSlot,
	}
}

func (q *qmi) Connect() error {
	cDevice := C.CString(q.device)
	defer C.free(unsafe.Pointer(cDevice))

	if C.libeuicc_qmi_apdu_connect(cDevice, C.int(q.uimSlot)) == -1 {
		return errors.New("failed to connect to QMI")
	}
	return nil
}

func (q *qmi) Disconnect() error {
	C.libeuicc_qmi_apdu_disconnect()
	return nil
}

func (q *qmi) Transmit(command []byte) ([]byte, error) {
	cCommand := C.CBytes(command)
	var cResponse *C.uint8_t
	var cResponseLen C.uint32_t
	defer C.free(unsafe.Pointer(cCommand))
	if C.libeuicc_qmi_apdu_transmit(&cResponse, &cResponseLen, (*C.uchar)(cCommand), C.uint(len(command))) == -1 {
		return nil, errors.New("failed to transmit APDU")
	}
	response := C.GoBytes(unsafe.Pointer(cResponse), C.int(cResponseLen))
	C.free(unsafe.Pointer(cResponse))
	return response, nil
}

func (q *qmi) OpenLogicalChannel(aid []byte) (int, error) {
	cAID := C.CBytes(aid)
	defer C.free(unsafe.Pointer(cAID))
	channel := C.libeuicc_qmi_apdu_open_logical_channel((*C.uchar)(cAID), C.uint8_t(len(aid)))
	if channel == -1 {
		return 0, errors.New("failed to open logical channel")
	}
	return int(channel), nil
}

func (q *qmi) CloseLogicalChannel(channel int) error {
	if C.libeuicc_qmi_apdu_close_logical_channel(C.uint8_t(channel)) == -1 {
		return errors.New("failed to close logical channel")
	}
	return nil
}
