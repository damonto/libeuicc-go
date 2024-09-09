package libeuicc

/*
#include <stdint.h>
#include <stdlib.h>

#include "interface.h"

extern int libeuiccApduConnect(struct euicc_ctx *ctx);
extern void libeuiccApduDisconnect(struct euicc_ctx *ctx);
extern int libeuiccApduOpenLogicalChannel(struct euicc_ctx *ctx, uint8_t *aid, uint8_t aid_len);
extern void libeuiccApduCloseLogicalChannel(struct euicc_ctx *ctx, uint8_t channel);
extern int libeuiccApduTransmit(struct euicc_ctx *ctx, uint8_t **rx, uint32_t *rx_len, uint8_t *tx, uint32_t tx_len);

static int libeuicc_forward_open_logical_channel(struct euicc_ctx *ctx, const uint8_t *aid, uint8_t aid_len) { return libeuiccApduOpenLogicalChannel(ctx, (uint8_t *)aid, aid_len); }
static int libeuicc_forward_apdu_transmit(struct euicc_ctx *ctx, uint8_t **rx, uint32_t *rx_len, const uint8_t *tx, uint32_t tx_len) { return libeuiccApduTransmit(ctx, rx, rx_len, (uint8_t *)tx, tx_len); }

static struct euicc_apdu_interface *init_apdu_interface()
{
	struct euicc_apdu_interface *apdu = (struct euicc_apdu_interface *)malloc(sizeof(struct euicc_apdu_interface));

	apdu->connect = libeuiccApduConnect;
	apdu->disconnect = libeuiccApduDisconnect;
	apdu->logic_channel_open = libeuicc_forward_open_logical_channel;
	apdu->logic_channel_close = libeuiccApduCloseLogicalChannel;
	apdu->transmit = libeuicc_forward_apdu_transmit;

	return apdu;
}
*/
import "C"
import "unsafe"

type APDU interface {
	Connect() error
	Disconnect() error
	Transmit(command []byte) ([]byte, error)
	OpenLogicalChannel(aid []byte) (int, error)
	CloseLogicalChannel(channel []byte) error
}

type uContext struct {
	driver APDU
}

func (e *Libeuicc) initAPDU(driver APDU) {
	e.euiccCtx.userdata = unsafe.Pointer(&uContext{
		driver: driver,
	})
	e.euiccCtx.apdu._interface = C.init_apdu_interface()
}

//export libeuiccApduConnect
func libeuiccApduConnect(ctx *C.struct_euicc_ctx) C.int {
	if (*uContext)(ctx.userdata).driver.Connect() != nil {
		return CError
	}
	return COK
}

//export libeuiccApduDisconnect
func libeuiccApduDisconnect(ctx *C.struct_euicc_ctx) {
	(*uContext)(ctx.userdata).driver.Disconnect()
}

//export libeuiccApduOpenLogicalChannel
func libeuiccApduOpenLogicalChannel(ctx *C.struct_euicc_ctx, aid *C.uint8_t, aid_len C.uint8_t) C.int {
	b := C.GoBytes(unsafe.Pointer(aid), C.int(aid_len))
	channel, err := (*uContext)(ctx.userdata).driver.OpenLogicalChannel(b)
	if err != nil {
		return CError
	}
	return C.int(channel)
}

//export libeuiccApduCloseLogicalChannel
func libeuiccApduCloseLogicalChannel(ctx *C.struct_euicc_ctx, channel C.uint8_t) {
	b := C.GoBytes(unsafe.Pointer(&channel), C.int(1))
	(*uContext)(ctx.userdata).driver.CloseLogicalChannel(b)
}

//export libeuiccApduTransmit
func libeuiccApduTransmit(ctx *C.struct_euicc_ctx, rx **C.uint8_t, rx_len *C.uint32_t, tx *C.uint8_t, tx_len C.uint32_t) C.int {
	b := C.GoBytes(unsafe.Pointer(tx), C.int(tx_len))
	r, err := (*uContext)(ctx.userdata).driver.Transmit(b)
	if err != nil {
		return CError
	}
	*rx = (*C.uint8_t)(C.CBytes(r))
	*rx_len = C.uint32_t(len(r))
	return COK
}
