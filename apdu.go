package libeuicc

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#include "interface.h"

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
	// Connect connects to the APDU interface. This is called before any other APDU operation.
	Connect() error
	// Disconnect disconnects from the APDU interface. This is called after the command execution is completed.
	Disconnect() error
	// Transmit sends the command to the APDU interface and returns the response.
	Transmit(command []byte) ([]byte, error)
	// OpenLogicalChannel opens a logical channel with the given AID and returns the channel number.
	OpenLogicalChannel(aid []byte) (int, error)
	// CloseLogicalChannel closes the logical channel with the given channel number.
	CloseLogicalChannel(channel int) error
}

func (e *Libeuicc) initAPDU() error {
	e.euiccCtx.apdu._interface = (*C.struct_euicc_apdu_interface)(C.malloc(C.sizeof_struct_euicc_apdu_interface))
	if e.euiccCtx.apdu._interface == nil {
		return errors.New("failed to allocate memory for APDU interface")
	}
	C.memset(unsafe.Pointer(e.euiccCtx.apdu._interface), 0, C.sizeof_struct_euicc_apdu_interface)
	e.euiccCtx.apdu._interface.userdata = unsafe.Pointer(e.driver)
	C.libeuicc_init_apdu_interface(e.euiccCtx.apdu._interface)
	return nil
}

//export libeuiccApduConnect
func libeuiccApduConnect(ctx *C.struct_euicc_ctx) C.int {
	if err := (*driver)(ctx.apdu._interface.userdata).apdu.Connect(); err != nil {
		logger.Error("APDU connect failed", err)
		return CError
	}
	logger.Debug("APDU connect success")
	return CSuccess
}

//export libeuiccApduDisconnect
func libeuiccApduDisconnect(ctx *C.struct_euicc_ctx) {
	if err := (*driver)(ctx.apdu._interface.userdata).apdu.Disconnect(); err != nil {
		logger.Error("APDU disconnect failed", err)
	}
	logger.Debug("APDU disconnect success")
}

//export libeuiccApduOpenLogicalChannel
func libeuiccApduOpenLogicalChannel(ctx *C.struct_euicc_ctx, aid *C.uint8_t, aid_len C.uint8_t) C.int {
	b := C.GoBytes(unsafe.Pointer(aid), C.int(aid_len))
	channel, err := (*driver)(ctx.apdu._interface.userdata).apdu.OpenLogicalChannel(b)
	if err != nil {
		logger.Error("APDU open logical channel failed", err)
		return CError
	}
	logger.Debug("APDU open logical channel success", "channel", channel)
	return C.int(channel)
}

//export libeuiccApduCloseLogicalChannel
func libeuiccApduCloseLogicalChannel(ctx *C.struct_euicc_ctx, channel C.uint8_t) {
	err := (*driver)(ctx.apdu._interface.userdata).apdu.CloseLogicalChannel(int(channel))
	if err != nil {
		logger.Error("APDU close logical channel failed", err, "channel", channel)
	}
	logger.Debug("APDU close logical channel success", "channel", channel)
}

//export libeuiccApduTransmit
func libeuiccApduTransmit(ctx *C.struct_euicc_ctx, rx **C.uint8_t, rx_len *C.uint32_t, tx *C.uint8_t, tx_len C.uint32_t) C.int {
	b := C.GoBytes(unsafe.Pointer(tx), C.int(tx_len))
	r, err := (*driver)(ctx.apdu._interface.userdata).apdu.Transmit(b)
	if err != nil {
		logger.Error("APDU transmit failed", err, "command", hex.EncodeToString(b))
		return CError
	}
	logger.Debug("APDU transmit success", "command", hex.EncodeToString(b), "response", hex.EncodeToString(r))
	*rx = (*C.uint8_t)(C.CBytes(r))
	*rx_len = C.uint32_t(len(r))
	return CSuccess
}
