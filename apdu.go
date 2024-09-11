package libeuicc

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#include "interface.h"

static struct qmi_data *qmi_priv;

extern int libeuiccApduConnect(struct euicc_ctx *ctx);
extern void libeuiccApduDisconnect(struct euicc_ctx *ctx);
extern int libeuiccApduOpenLogicalChannel(struct euicc_ctx *ctx, uint8_t *aid, uint8_t aid_len);
extern void libeuiccApduCloseLogicalChannel(struct euicc_ctx *ctx, uint8_t channel);
extern int libeuiccApduTransmit(struct euicc_ctx *ctx, uint8_t **rx, uint32_t *rx_len, uint8_t *tx, uint32_t tx_len);

static int libeuicc_forward_open_logical_channel(struct euicc_ctx *ctx, const uint8_t *aid, uint8_t aid_len)
{
	return libeuiccApduOpenLogicalChannel(ctx, (uint8_t *)aid, aid_len);
}

static int libeuicc_forward_apdu_transmit(struct euicc_ctx *ctx, uint8_t **rx, uint32_t *rx_len, const uint8_t *tx, uint32_t tx_len)
{
	return libeuiccApduTransmit(ctx, rx, rx_len, (uint8_t *)tx, tx_len);
}

static void libeuicc_init_apdu_interface(struct euicc_apdu_interface *ifstruct)
{
	ifstruct->connect = libeuiccApduConnect;
	ifstruct->disconnect = libeuiccApduDisconnect;
	ifstruct->logic_channel_open = libeuicc_forward_open_logical_channel;
	ifstruct->logic_channel_close = libeuiccApduCloseLogicalChannel;
	ifstruct->transmit = libeuicc_forward_apdu_transmit;
}
*/
import "C"
import (
	"encoding/hex"
	"errors"
	"unsafe"
)

type APDU interface {
	Connect() error
	Disconnect() error
	Transmit(command []byte) ([]byte, error)
	OpenLogicalChannel(aid []byte) (int, error)
	CloseLogicalChannel(channel int) error
}

type apduCtx struct {
	driver APDU
}

func (e *Libeuicc) initAPDU(driver APDU) error {
	e.ctx.apdu._interface = (*C.struct_euicc_apdu_interface)(C.malloc(C.sizeof_struct_euicc_apdu_interface))
	if e.ctx.apdu._interface == nil {
		return errors.New("failed to allocate memory for APDU interface")
	}
	C.memset(unsafe.Pointer(e.ctx.apdu._interface), 0, C.sizeof_struct_euicc_apdu_interface)
	e.ctx.apdu._interface.userdata = unsafe.Pointer(&apduCtx{driver: driver})
	C.libeuicc_init_apdu_interface(e.ctx.apdu._interface)
	return nil
}

//export libeuiccApduConnect
func libeuiccApduConnect(ctx *C.struct_euicc_ctx) C.int {
	if driver, ok := (*apduCtx)(ctx.apdu._interface.userdata).driver.(interface{ Connect() error }); ok {
		if err := driver.Connect(); err != nil {
			logger.Error("APDU connect failed", err)
			return CError
		}
		logger.Debug("APDU connect success")
		return COK
	}
	return CError
}

//export libeuiccApduDisconnect
func libeuiccApduDisconnect(ctx *C.struct_euicc_ctx) {
	if driver, ok := (*apduCtx)(ctx.apdu._interface.userdata).driver.(interface{ Disconnect() error }); ok {
		if err := driver.Disconnect(); err != nil {
			logger.Error("APDU disconnect failed", err)
		}
		logger.Debug("APDU disconnect success")
	}
}

//export libeuiccApduOpenLogicalChannel
func libeuiccApduOpenLogicalChannel(ctx *C.struct_euicc_ctx, aid *C.uint8_t, aid_len C.uint8_t) C.int {
	if driver, ok := (*apduCtx)(ctx.apdu._interface.userdata).driver.(interface{ OpenLogicalChannel(aid []byte) (int, error) }); ok {
		b := C.GoBytes(unsafe.Pointer(aid), C.int(aid_len))
		channel, err := driver.OpenLogicalChannel(b)
		if err != nil {
			logger.Error("APDU open logical channel failed", err)
			return CError
		}
		logger.Debug("APDU open logical channel success", "channel", channel)
		return C.int(channel)
	}
	return CError
}

//export libeuiccApduCloseLogicalChannel
func libeuiccApduCloseLogicalChannel(ctx *C.struct_euicc_ctx, channel C.uint8_t) {
	if channel <= 0 {
		return
	}
	if driver, ok := (*apduCtx)(ctx.apdu._interface.userdata).driver.(interface{ CloseLogicalChannel(channel int) error }); ok {
		if err := driver.CloseLogicalChannel(int(channel)); err != nil {
			logger.Error("APDU close logical channel failed", err, "channel", channel)
		} else {
			logger.Debug("APDU close logical channel success", "channel", channel)
		}
	}
}

//export libeuiccApduTransmit
func libeuiccApduTransmit(ctx *C.struct_euicc_ctx, rx **C.uint8_t, rx_len *C.uint32_t, tx *C.uint8_t, tx_len C.uint32_t) C.int {
	if driver, ok := (*apduCtx)(ctx.apdu._interface.userdata).driver.(interface {
		Transmit(command []byte) ([]byte, error)
	}); ok {
		b := C.GoBytes(unsafe.Pointer(tx), C.int(tx_len))
		r, err := driver.Transmit(b)
		if err != nil {
			logger.Error("APDU transmit failed", err, "command", hex.EncodeToString(b))
			return CError
		}
		logger.Debug("APDU transmit success", "command", hex.EncodeToString(b), "response", hex.EncodeToString(r))
		*rx = (*C.uint8_t)(C.CBytes(r))
		*rx_len = C.uint32_t(len(r))
		return COK
	}
	return CError
}
