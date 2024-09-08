package main

/*
#include <stdint.h>
#include <stdlib.h>

#include "euicc.h"
#include "interface.h"

extern int httpTransmit(struct euicc_ctx *ctx, char *url, uint32_t *rcode, uint8_t **rx, uint32_t *rx_len, uint8_t *tx, uint32_t tx_len, char **headers);

static int g_http_transmit(struct euicc_ctx *ctx, const char *url, uint32_t *rcode, uint8_t **rx, uint32_t *rx_len, const uint8_t *tx, uint32_t tx_len, const char **headers) {
	return httpTransmit(ctx, (char *)url, rcode, rx, rx_len, (uint8_t *)tx, tx_len, (char **)headers);
};

static struct euicc_http_interface *init_http_interface() {
	struct euicc_http_interface *http = (struct euicc_http_interface *)malloc(sizeof(struct euicc_http_interface));

	http->transmit = g_http_transmit;

	return http;
}
*/
import "C"
import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"strings"
	"time"
	"unsafe"
)

var certs = map[string]string{
	"GSM Association - RSP2 Root CI1": `-----BEGIN CERTIFICATE-----
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

func initHttp(ctx *C.struct_euicc_ctx) {
	ctx.http._interface = C.init_http_interface()
}

//export httpTransmit
func httpTransmit(ctx *C.struct_euicc_ctx, url *C.char, rcode *C.uint32_t, rx **C.uint8_t, rx_len *C.uint32_t, tx *C.uint8_t, tx_len C.uint32_t, headers **C.char) C.int {
	req, err := http.NewRequest("POST", C.GoString(url), bytes.NewBuffer(C.GoBytes(unsafe.Pointer(tx), C.int(tx_len))))
	if err != nil {
		return C.int(-1)
	}
	req.Header.Set("Content-Type", "application/json")
	for _, header := range GoStrings(headers) {
		kv := strings.SplitN(header, ":", 2)
		req.Header.Add(kv[0], kv[1])
	}

	http.DefaultClient.Timeout = 60 * time.Second
	rootCAs := x509.NewCertPool()
	for _, cert := range certs {
		rootCAs.AppendCertsFromPEM([]byte(cert))
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		RootCAs: rootCAs,
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return C.int(-1)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	*rx = (*C.uint8_t)(C.CBytes(body))
	*rx_len = C.uint32_t(len(body))
	*rcode = C.uint32_t(resp.StatusCode)
	return C.int(0)
}
