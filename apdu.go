package main

/*
#include <stdint.h>
#include <stdlib.h>

#include "interface.h"

extern int connect(struct euicc_ctx *ctx);
extern void disconnect(struct euicc_ctx *ctx);
extern int openLogicalChannel(struct euicc_ctx *ctx, uint8_t *aid, uint8_t aid_len);
extern void closeLogicalChannel(struct euicc_ctx *ctx, uint8_t channel);
extern int apduTransmit(struct euicc_ctx *ctx, uint8_t **rx, uint32_t *rx_len, uint8_t *tx, uint32_t tx_len);

static int g_open_logical_channel(struct euicc_ctx *ctx, const uint8_t *aid, uint8_t aid_len) { return openLogicalChannel(ctx, (uint8_t *)aid, aid_len); }
static int g_apdu_transmit(struct euicc_ctx *ctx, uint8_t **rx, uint32_t *rx_len, const uint8_t *tx, uint32_t tx_len) { return apduTransmit(ctx, rx, rx_len, (uint8_t *)tx, tx_len); }

static struct euicc_apdu_interface *init_apdu_interface() {
	struct euicc_apdu_interface *apdu = (struct euicc_apdu_interface *)malloc(sizeof(struct euicc_apdu_interface));

	apdu->connect = connect;
	apdu->disconnect = disconnect;
	apdu->logic_channel_open = g_open_logical_channel;
	apdu->logic_channel_close = closeLogicalChannel;
	apdu->transmit = g_apdu_transmit;

	return apdu;
}
*/
import "C"
import (
	"unsafe"
)

type APDU interface {
	Connect() error
	Disconnect() error
	Transmit(command []byte) ([]byte, error)
	OpenLogicalChannel(aid []byte) (int, error)
	CloseLogicalChannel(channel []byte) error
}

var apdu APDU

func initAPDU(ctx *C.struct_euicc_ctx, driver APDU) {
	apdu = driver
	ctx.apdu._interface = C.init_apdu_interface()
}

//export connect
func connect(ctx *C.struct_euicc_ctx) C.int {
	if apdu.Connect() != nil {
		return CError
	}
	return COK
}

//export disconnect
func disconnect(ctx *C.struct_euicc_ctx) {
	apdu.Disconnect()
}

//export openLogicalChannel
func openLogicalChannel(ctx *C.struct_euicc_ctx, aid *C.uint8_t, aid_len C.uint8_t) C.int {
	b := C.GoBytes(unsafe.Pointer(aid), C.int(aid_len))
	channel, err := apdu.OpenLogicalChannel(b)
	if err != nil {
		return CError
	}
	return C.int(channel)
}

//export closeLogicalChannel
func closeLogicalChannel(ctx *C.struct_euicc_ctx, channel C.uint8_t) {
	b := C.GoBytes(unsafe.Pointer(&channel), C.int(1))
	apdu.CloseLogicalChannel(b)
}

//export apduTransmit
func apduTransmit(ctx *C.struct_euicc_ctx, rx **C.uint8_t, rx_len *C.uint32_t, tx *C.uint8_t, tx_len C.uint32_t) C.int {
	b := C.GoBytes(unsafe.Pointer(tx), C.int(tx_len))
	r, err := apdu.Transmit(b)
	if err != nil {
		return CError
	}
	*rx = (*C.uint8_t)(C.CBytes(r))
	*rx_len = C.uint32_t(len(r))
	return COK
}
