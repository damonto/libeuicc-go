package libeuicc

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#include "euicc.h"
#include "interface.h"

extern int libeuiccHttpTransmit(struct euicc_ctx *ctx, char *url, uint32_t *rcode, uint8_t **rx, uint32_t *rx_len, uint8_t *tx, uint32_t tx_len, char **headers);

static int libeuicc_forward_http_transmit(struct euicc_ctx *ctx, const char *url, uint32_t *rcode, uint8_t **rx, uint32_t *rx_len, const uint8_t *tx, uint32_t tx_len, const char **headers)
{
	return libeuiccHttpTransmit(ctx, (char *)url, rcode, rx, rx_len, (uint8_t *)tx, tx_len, (char **)headers);
}

static void libeuicc_init_http_interface(struct euicc_http_interface *ifstruct)
{
	ifstruct->transmit = libeuicc_forward_http_transmit;
}
*/
import "C"
import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
	"unsafe"
)

var certs = []string{
	// GSM Association - RSP2 Root CI1
	`-----BEGIN CERTIFICATE-----
MIICSTCCAe+gAwIBAgIQbmhWeneg7nyF7hg5Y9+qejAKBggqhkjOPQQDAjBEMRgw
FgYDVQQKEw9HU00gQXNzb2NpYXRpb24xKDAmBgNVBAMTH0dTTSBBc3NvY2lhdGlv
biAtIFJTUDIgUm9vdCBDSTEwIBcNMTcwMjIyMDAwMDAwWhgPMjA1MjAyMjEyMzU5
NTlaMEQxGDAWBgNVBAoTD0dTTSBBc3NvY2lhdGlvbjEoMCYGA1UEAxMfR1NNIEFz
c29jaWF0aW9uIC0gUlNQMiBSb290IENJMTBZMBMGByqGSM49AgEGCCqGSM49AwEH
A0IABJ1qutL0HCMX52GJ6/jeibsAqZfULWj/X10p/Min6seZN+hf5llovbCNuB2n
unLz+O8UD0SUCBUVo8e6n9X1TuajgcAwgb0wDgYDVR0PAQH/BAQDAgEGMA8GA1Ud
EwEB/wQFMAMBAf8wEwYDVR0RBAwwCogIKwYBBAGC6WAwFwYDVR0gAQH/BA0wCzAJ
BgdngRIBAgEAME0GA1UdHwRGMEQwQqBAoD6GPGh0dHA6Ly9nc21hLWNybC5zeW1h
dXRoLmNvbS9vZmZsaW5lY2EvZ3NtYS1yc3AyLXJvb3QtY2kxLmNybDAdBgNVHQ4E
FgQUgTcPUSXQsdQI1MOyMubSXnlb6/swCgYIKoZIzj0EAwIDSAAwRQIgIJdYsOMF
WziPK7l8nh5mu0qiRiVf25oa9ullG/OIASwCIQDqCmDrYf+GziHXBOiwJwnBaeBO
aFsiLzIEOaUuZwdNUw==
-----END CERTIFICATE-----`,
}

func (e *Libeuicc) initHttp() error {
	e.euiccCtx.http._interface = (*C.struct_euicc_http_interface)(C.malloc(C.sizeof_struct_euicc_http_interface))
	if e.euiccCtx.http._interface == nil {
		return errors.New("failed to allocate memory for http interface")
	}
	C.memset(unsafe.Pointer(e.euiccCtx.http._interface), 0, C.sizeof_struct_euicc_http_interface)
	C.libeuicc_init_http_interface(e.euiccCtx.http._interface)
	return nil
}

//export libeuiccHttpTransmit
func libeuiccHttpTransmit(_ *C.struct_euicc_ctx, url *C.char, rcode *C.uint32_t, rx **C.uint8_t, rx_len *C.uint32_t, tx *C.uint8_t, tx_len C.uint32_t, headers **C.char) C.int {
	r, err := http.NewRequest(http.MethodPost, C.GoString(url), bytes.NewBuffer(C.GoBytes(unsafe.Pointer(tx), C.int(tx_len))))
	if err != nil {
		logger.Error("Failed to create http request", err)
		return CError
	}

	for _, header := range GoStrings(headers) {
		kv := strings.SplitN(header, ":", 2)
		r.Header.Add(kv[0], kv[1])
	}

	rootCAs := x509.NewCertPool()
	for _, cert := range certs {
		rootCAs.AppendCertsFromPEM([]byte(cert))
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: rootCAs,
			},
		},
	}

	resp, err := client.Do(r)
	if err != nil {
		logger.Error("Failed to send http request", err, "url", r.URL.String())
		return CError
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	logger.Debug("Http transmit success", "url", r.URL.String(), "status", resp.StatusCode, "body", string(body))

	*rx = (*C.uint8_t)(C.CBytes(body))
	*rx_len = C.uint32_t(len(body))
	*rcode = C.uint32_t(resp.StatusCode)
	return CSuccess
}
