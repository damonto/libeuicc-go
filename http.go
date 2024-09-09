package libeuicc

/*
#include <stdint.h>
#include <stdlib.h>

#include "euicc.h"
#include "interface.h"

extern int libeuiccHttpTransmit(struct euicc_ctx *ctx, char *url, uint32_t *rcode, uint8_t **rx, uint32_t *rx_len, uint8_t *tx, uint32_t tx_len, char **headers);

static int libeuicc_forward_http_transmit(struct euicc_ctx *ctx, const char *url, uint32_t *rcode, uint8_t **rx, uint32_t *rx_len, const uint8_t *tx, uint32_t tx_len, const char **headers) {
	return libeuiccHttpTransmit(ctx, (char *)url, rcode, rx, rx_len, (uint8_t *)tx, tx_len, (char **)headers);
};

static struct euicc_http_interface *init_http_interface()
{
	struct euicc_http_interface *http = (struct euicc_http_interface *)malloc(sizeof(struct euicc_http_interface));

	http->transmit = libeuicc_forward_http_transmit;

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
	// OISTE GSMA CI G1
	`-----BEGIN CERTIFICATE-----
MIIB9zCCAZ2gAwIBAgIUSpBSCCDYPOEG/IFHUCKpZ2pIAQMwCgYIKoZIzj0EAwIw
QzELMAkGA1UEBhMCQ0gxGTAXBgNVBAoMEE9JU1RFIEZvdW5kYXRpb24xGTAXBgNV
BAMMEE9JU1RFIEdTTUEgQ0kgRzEwIBcNMjQwMTE2MjMxNzM5WhgPMjA1OTAxMDcy
MzE3MzhaMEMxCzAJBgNVBAYTAkNIMRkwFwYDVQQKDBBPSVNURSBGb3VuZGF0aW9u
MRkwFwYDVQQDDBBPSVNURSBHU01BIENJIEcxMFkwEwYHKoZIzj0CAQYIKoZIzj0D
AQcDQgAEvZ3s3PFC4NgrCcCMmHJ6DJ66uzAHuLcvjJnOn+TtBNThS7YHLDyHCa2v
7D+zTP+XTtgqgcLoB56Gha9EQQQ4xKNtMGswDwYDVR0TAQH/BAUwAwEB/zAQBgNV
HREECTAHiAVghXQFDjAXBgNVHSABAf8EDTALMAkGB2eBEgECAQAwHQYDVR0OBBYE
FEwnlnrSDBSzkelgHkHmBK1XwCIvMA4GA1UdDwEB/wQEAwIBBjAKBggqhkjOPQQD
AgNIADBFAiBVcywTj017jKpAQ+gwy4MqK2hQvzve6lkvQkgSP6ykHwIhAI0KFwCD
jnPbmcJsG41hUrWNlf+IcrMvFuYii0DasBNi
-----END CERTIFICATE-----`,
	// Symantec Corporation RSP Test Root CA - For Test Purposes Only
	`-----BEGIN CERTIFICATE-----
MIICkDCCAjagAwIBAgIQPfCO5OYL+cdbbx2ETDO7DDAKBggqhkjOPQQDAjBoMR0w
GwYDVQQKExRTeW1hbnRlYyBDb3Jwb3JhdGlvbjFHMEUGA1UEAxM+U3ltYW50ZWMg
Q29ycG9yYXRpb24gUlNQIFRlc3QgUm9vdCBDQSAtIEZvciBUZXN0IFB1cnBvc2Vz
IE9ubHkwHhcNMTcwNzExMDAwMDAwWhcNNDkxMjMxMjM1OTU5WjBoMR0wGwYDVQQK
ExRTeW1hbnRlYyBDb3Jwb3JhdGlvbjFHMEUGA1UEAxM+U3ltYW50ZWMgQ29ycG9y
YXRpb24gUlNQIFRlc3QgUm9vdCBDQSAtIEZvciBUZXN0IFB1cnBvc2VzIE9ubHkw
WTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQlbEYt9PTmdWcaX5WC68SYTFyZcbBN
vFpJW6bZQpERlMIAuzEpgscbTDccHtNpDqJwMqZXCO7ebCmRLyI6jqe3o4HBMIG+
MA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MBcGA1UdIAEB/wQNMAsw
CQYHZ4ESAQIBADBPBgNVHR8ESDBGMESgQqBAhj5odHRwOi8vcGtpLWNybC5zeW1h
dXRoLmNvbS9TeW1hbnRlY1JTUFRlc3RSb290Q0EvTGF0ZXN0Q1JMLmNybDASBgNV
HREECzAJiAcrBgEEAYMJMB0GA1UdDgQWBBRmWhQz1nwaLF24tSyWfxCgV7pcsjAK
BggqhkjOPQQDAgNIADBFAiAQ1quTqcexvDnKvmAkqoQP09QMXAXxlCyma82NtrYq
UQIhAP/W6pRamBGhSliV+EancgbZj+VoOkKdj0o7sP/cKdhZ
-----END CERTIFICATE-----`,
	// GSMA Test CI (SGP.26 v1)
	`-----BEGIN CERTIFICATE-----
MIICUDCCAfegAwIBAgIJALh086v6bETTMAoGCCqGSM49BAMCMEQxEDAOBgNVBAMM
B1Rlc3QgQ0kxETAPBgNVBAsMCFRFU1RDRVJUMRAwDgYDVQQKDAdSU1BURVNUMQsw
CQYDVQQGEwJJVDAgFw0yMDA0MDEwODI3NTFaGA8yMDU1MDQwMTA4Mjc1MVowRDEQ
MA4GA1UEAwwHVGVzdCBDSTERMA8GA1UECwwIVEVTVENFUlQxEDAOBgNVBAoMB1JT
UFRFU1QxCzAJBgNVBAYTAklUMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAElAZX
pnPcKI+J1S6opHcEmSeR+cNLADbmM+LQy6lFTWXbMusXmBeZ0vJDiO4rlcEJRUbJ
eQHOrrqWUJGaLiDSKaOBzzCBzDAdBgNVHQ4EFgQU9UFyvfmKldZcvriKOKHBHYAK
hcMwDwYDVR0TAQH/BAUwAwEB/zAXBgNVHSABAf8EDTALMAkGB2eBEgECAQAwDgYD
VR0PAQH/BAQDAgEGMA4GA1UdEQQHMAWIA4g3ATBhBgNVHR8EWjBYMCqgKKAmhiRo
dHRwOi8vY2kudGVzdC5leGFtcGxlLmNvbS9DUkwtQS5jcmwwKqAooCaGJGh0dHA6
Ly9jaS50ZXN0LmV4YW1wbGUuY29tL0NSTC1CLmNybDAKBggqhkjOPQQDAgNHADBE
AiBSdWqvwgIKbOy/Ll88IIklEP8pdR0pi9OwFdlgWk/mfQIgV5goNuTSBd3S5sPB
tFWTf2tuSTtgL9G2bDV0iak192s=
-----END CERTIFICATE-----`,
	// GSMA Test CI (SGP.26 v1, BRP P256r1)
	`-----BEGIN CERTIFICATE-----
MIICUTCCAfigAwIBAgIJALh086v6bETTMAoGCCqGSM49BAMCMEQxEDAOBgNVBAMM
B1Rlc3QgQ0kxETAPBgNVBAsMCFRFU1RDRVJUMRAwDgYDVQQKDAdSU1BURVNUMQsw
CQYDVQQGEwJJVDAgFw0yMDA0MDEwODI3NTFaGA8yMDU1MDQwMTA4Mjc1MVowRDEQ
MA4GA1UEAwwHVGVzdCBDSTERMA8GA1UECwwIVEVTVENFUlQxEDAOBgNVBAoMB1JT
UFRFU1QxCzAJBgNVBAYTAklUMFowFAYHKoZIzj0CAQYJKyQDAwIIAQEHA0IABCeH
tNVu2CSp5r4E4Yh/a5i6/rjHY/UoN/cBE+k2Tt2+E5vAx95+Fo8eXNDBhTT8UGTm
T2htxTMnyn8dzqhaKZSjgc8wgcwwHQYDVR0OBBYEFMC8cLo2kp1DtGf/V1cFMOV6
uPzYMA8GA1UdEwEB/wQFMAMBAf8wFwYDVR0gAQH/BA0wCzAJBgdngRIBAgEAMA4G
A1UdDwEB/wQEAwIBBjAOBgNVHREEBzAFiAOINwEwYQYDVR0fBFowWDAqoCigJoYk
aHR0cDovL2NpLnRlc3QuZXhhbXBsZS5jb20vQ1JMLUEuY3JsMCqgKKAmhiRodHRw
Oi8vY2kudGVzdC5leGFtcGxlLmNvbS9DUkwtQi5jcmwwCgYIKoZIzj0EAwIDRwAw
RAIgPYrf0CKl0FBMUaHx5xS1duTDbQ4wBZN3qKBeNniuux0CIHBek2vLfoANAdtt
f5u5Ce6DVC2oIfpn5UnS24F3oMqM
-----END CERTIFICATE-----`,
	// GSMA Test CI (SGP.26 v3)
	`-----BEGIN CERTIFICATE-----
MIIB5DCCAYugAwIBAgIBADAKBggqhkjOPQQDAjBEMRAwDgYDVQQDDAdUZXN0IENJ
MREwDwYDVQQLDAhURVNUQ0VSVDEQMA4GA1UECgwHUlNQVEVTVDELMAkGA1UEBhMC
SVQwIBcNMjMwNjAyMTMwNTQzWhgPMjA1ODA2MDExMzA1NDNaMEQxEDAOBgNVBAMM
B1Rlc3QgQ0kxETAPBgNVBAsMCFRFU1RDRVJUMRAwDgYDVQQKDAdSU1BURVNUMQsw
CQYDVQQGEwJJVDBaMBQGByqGSM49AgEGCSskAwMCCAEBBwNCAASF7cCXanl/xSJe
PwIeEUeZk4zPPM3iE16JbpOWPqPXaJwGmMKvHwQlRxiLtPWrRBalgkzrr4RgYIqD
aTcnvxoFo2swaTAdBgNVHQ4EFgQUIgn2HNnsXJyFTnhzQf+D7Pl3alswDgYDVR0P
AQH/BAQDAgEGMBcGA1UdIAEB/wQNMAswCQYHZ4ESAQIBADAPBgNVHRMBAf8EBTAD
AQH/MA4GA1UdEQQHMAWIA4g3ATAKBggqhkjOPQQDAgNHADBEAiBLLHbhrIvy1Cue
7lDUlQZY2EOK7/I/o2CQO0pj76OqzQIgTQ+kE02RPbMuflDbXKRuVDKFvfZ/vHEW
QKvBPWehIXI=
-----END CERTIFICATE-----`,
	// GSMA Test CI (SGP.26 v3, BRP P256r1)
	`-----BEGIN CERTIFICATE-----
MIIB4zCCAYqgAwIBAgIBADAKBggqhkjOPQQDAjBEMRAwDgYDVQQDDAdUZXN0IENJ
MREwDwYDVQQLDAhURVNUQ0VSVDEQMA4GA1UECgwHUlNQVEVTVDELMAkGA1UEBhMC
SVQwIBcNMjMwNTMxMTI1MDI4WhgPMjA1ODA1MzAxMjUwMjhaMEQxEDAOBgNVBAMM
B1Rlc3QgQ0kxETAPBgNVBAsMCFRFU1RDRVJUMRAwDgYDVQQKDAdSU1BURVNUMQsw
CQYDVQQGEwJJVDBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABLlQW4kHaMJSrAK4
nVKjGIgKWYxick+Y1x0MKO/Bsb3+KxMdnAObkPZjLosKlKCnH2bHUHhqRyDDSc2Y
9+wB6A6jazBpMB0GA1UdDgQWBBQ07s8TFWUY1I0wvfBoU0BNEV+VXTAOBgNVHQ8B
Af8EBAMCAQYwFwYDVR0gAQH/BA0wCzAJBgdngRIBAgEAMA8GA1UdEwEB/wQFMAMB
Af8wDgYDVR0RBAcwBYgDiDcBMAoGCCqGSM49BAMCA0cAMEQCIEuYVB+bwdn5Z6sL
eKFS07FnvHY03QqDm8XYxdjDAxZuAiBneNr+fBYeqDulQWfrGXFLDTbsFBENNdDj
jvcHgHpATQ==
-----END CERTIFICATE-----`,
	// Taier eSIM Root CA - NISTP256
	`-----BEGIN CERTIFICATE-----
MIIC5DCCAoqgAwIBAgIDAZ9KMAoGCCqGSM49BAMCMHgxCzAJBgNVBAYTAkNOMRAw
DgYDVQQIDAdCRUlKSU5HMRAwDgYDVQQHDAdCRUlKSU5HMQ0wCwYDVQQLDARDVFRM
MQ4wDAYDVQQKDAVDQUlDVDEmMCQGA1UEAwwdVGFpZXIgZVNJTSBSb290IENBIC0g
TklTVFAyNTYwHhcNMjAwNDAzMDgxODEyWhcNMzAwNDAzMDgxODEyWjCBgzELMAkG
A1UEBhMCQ0gxIjAgBgkqhkiG9w0BCQEWE3poZW5naGFpeGlhQGNhdHIuY24xEDAO
BgNVBAgMB0JFSUpJTkcxDDAKBgNVBAsMA0NUQTENMAsGA1UECgwEQ1RUTDEhMB8G
A1UEAwwYd3d3LmVzaW10ZXN0LmNoaW5hdHRsLmNuMFkwEwYHKoZIzj0CAQYIKoZI
zj0DAQcDQgAEVZ4lJlh47idWviiKhQfuYG/JHs8XCxS29bTd5hIidqfsDZDgTe1c
Mf/Sv0+RUy9mQaLdb49vY+2HQWoO9SSu0KOB9jCB8zAfBgNVHSMEGDAWgBQUgDD8
JGwC/yIhBhJajWvamQr76TAdBgNVHQ4EFgQUEtoer6QN28gwv6clRAX5+bvIeLAw
DgYDVR0PAQH/BAQDAgeAMBcGA1UdIAEB/wQNMAswCQYHZ4ESAQIBBDAVBgNVHREE
DjAMiAorBgEEAYORY2UVMHEGA1UdHwRqMGgwMqAwoC6GLGh0dHA6Ly8xMTEuMjA0
LjE3Ni4yNTQ6MTg4ODkvZG93bmxvYWQvbjEuY3JsMDKgMKAuhixodHRwOi8vMTEx
LjIwNC4xNzYuMjU0OjE4ODg5L2Rvd25sb2FkL24yLmNybDAKBggqhkjOPQQDAgNI
ADBFAiBvFAodYVHsgZYzQjsEVlKmXo/eiP9LXutXjz3TK0otZgIhAJLPe0mXYIyu
y10edQdrERLpqsKMACkRNPvPpFu84Wmh
-----END CERTIFICATE-----`,
	// China Unicom eSIM Root CA
	`-----BEGIN CERTIFICATE-----
MIICxTCCAmqgAwIBAgIDAYbzMAoGCCqGSM49BAMCMHExCzAJBgNVBAYTAkNOMQ4w
DAYDVQQIEwVKSUxJTjESMBAGA1UEBxMJQ0hBTkdDSFVOMQ0wCwYDVQQLEwRDVUNB
MQswCQYDVQQKEwJDVTEiMCAGA1UEAxMZQ2hpbmEgVW5pY29tIGVTSU0gUm9vdCBD
QTAeFw0xNzAzMDYxNzAwMDBaFw0zNzAzMDYxNzAwMDBaMHExCzAJBgNVBAYTAkNO
MQ4wDAYDVQQIEwVKSUxJTjESMBAGA1UEBxMJQ0hBTkdDSFVOMQ0wCwYDVQQLEwRD
VUNBMQswCQYDVQQKEwJDVTEiMCAGA1UEAxMZQ2hpbmEgVW5pY29tIGVTSU0gUm9v
dCBDQTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABFUCQBaBtMGOyCIJ21+RZUZY
+n0IA4gaNW+M3s7WnzzdM0mqL6Mm+wkQAzxGbi9H0RcUu5YURJX9iI+wcb/LJQaj
gfAwge0wHQYDVR0OBBYEFBa10WBI4+oCvUtgbl93pL8ggI2DMA8GA1UdEwEB/wQF
MAMBAf8wFwYDVR0gAQH/BA0wCzAJBgdngRIBAgEAMHsGA1UdHwR0MHIwN6A1oDOG
MWh0dHA6Ly9lc2ltY3JsLnVuaS1jYS5jb20uY246NjY2Ni9kb3dubG9hZC9uMS5j
cmwwN6A1oDOGMWh0dHA6Ly9lc2ltY3JsLnVuaS1jYS5jb20uY246NjY2Ni9kb3du
bG9hZC9uMi5jcmwwFQYDVR0RBA4wDIgKKwYBBAGC9UZlAjAOBgNVHQ8BAf8EBAMC
AQYwCgYIKoZIzj0EAwIDSQAwRgIhAOjQpOCw6676m2dHeSCfmPbPdm/WFgYctXld
IHTvlvILAiEAjVnvej0qnMv//JPpz8CtpMvcg0xOuu7TAgATojT0DkM=
-----END CERTIFICATE-----`,
	// CMCA eSIM Root CA_NIST
	`-----BEGIN CERTIFICATE-----
MIICUzCCAfqgAwIBAgIDAYuUMAoGCCqGSM49BAMCMD0xCzAJBgNVBAYTAkNOMQ0w
CwYDVQQKDARDTUNBMR8wHQYDVQQDDBZDTUNBIGVTSU0gUm9vdCBDQV9OSVNUMCAX
DTE3MDkwNTAyMDAwMFoYDzIwNTIwOTA1MDIwMDAwWjA9MQswCQYDVQQGEwJDTjEN
MAsGA1UECgwEQ01DQTEfMB0GA1UEAwwWQ01DQSBlU0lNIFJvb3QgQ0FfTklTVDBZ
MBMGByqGSM49AgEGCCqGSM49AwEHA0IABDLai0jKDVz1o0G09mLR6UhfZGOHB48u
NJbtUob1ySQFJ4lbydJXmIEIxoasPnwgZPoXv6nxLc2SIUmLFIp9d6ijgeYwgeMw
HQYDVR0OBBYEFM320cCnsH+YqGG243i4L2SNmWY+MA8GA1UdEwEB/wQFMAMBAf8w
FwYDVR0gAQH/BA0wCzAJBgdngRIBAgEAMHEGA1UdHwRqMGgwMqAwoC6GLGh0dHA6
Ly8yMjEuMTc2LjY1LjU1OjgwODEvY3JsZG93bmxvYWQvbkEuY3JsMDKgMKAuhixo
dHRwOi8vMjIxLjE3Ni42NS41NTo4MDgxL2NybGRvd25sb2FkL25CLmNybDAVBgNV
HREEDjAMiAorBgEEAYOLKGUBMA4GA1UdDwEB/wQEAwIBBjAKBggqhkjOPQQDAgNH
ADBEAiBIVvu51/6PrNnDyf0KDrKDF00Ix8NlCsV7wG61MlsfNQIgNiIdw4sMnP+F
LyCiH8gY+DmBlwAFRQK00aUziQsG+78=
-----END CERTIFICATE-----`,
	// CCS NETCA eSIM ECC RootCA
	`-----BEGIN CERTIFICATE-----
MIICjzCCAjWgAwIBAgIUecXMKhar2+vsbdKKaC40399+H1kwCgYIKoZIzj0EAwIw
VzELMAkGA1UEBhMCQ04xJDAiBgNVBAoMG05FVENBIENlcnRpZmljYXRlIEF1dGhv
cml0eTEiMCAGA1UEAwwZQ0NTIE5FVENBIGVTSU0gRUNDIFJvb3RDQTAeFw0yMjEy
MjUxNjAwMDBaFw0yNDEyMzAxNjAwMDBaMHAxCzAJBgNVBAYTAkNOMRAwDgYDVQQI
DAdCRUlKSU5HMRAwDgYDVQQHDAdCRUlKSU5HMRYwFAYDVQQKDA1DaGluYSBUZWxl
Y29tMQswCQYDVQQLDAJJVDEYMBYGA1UEAwwPZXNpbS5jcm0uMTg5LmNuMFkwEwYH
KoZIzj0CAQYIKoZIzj0DAQcDQgAEZemAjELdmo90OrsF1TlkciCGiEktU0UBixPz
iGuNA9/XrNCwrGn7wndAlIPuY6at5MWoL+Jiu9LdBrLLR3vrHKOBxTCBwjAfBgNV
HSMEGDAWgBTT74P8UD+cbsS39q4FW38Tc9SrHjAdBgNVHQ4EFgQUqx9lfmqG34HZ
X/cOB1BB4bDTLOgwDgYDVR0PAQH/BAQDAgeAMBcGA1UdIAEB/wQNMAswCQYHZ4ES
AQIBBDAUBgNVHREEDTALiAkrBgEEAYODWQwwDAYDVR0TAQH/BAIwADAzBgNVHR8E
LDAqMCigJqAkhiJodHRwOi8vY3JsLmNuY2EubmV0L2VzaW0vY2NzL2EuY3JsMAoG
CCqGSM49BAMCA0gAMEUCIFrZxg+22xzDrLNH2WUd+sMvyBimnE8j5g7RCHhswixs
AiEA0euPG/TA5rEFNUQm5GnktCAIskKVR5GKRKa94JWVSkY=
-----END CERTIFICATE-----`,
	// Entrust eSIM Certification Authority
	`-----BEGIN CERTIFICATE-----
MIIC6DCCAo2gAwIBAgIRAIy4GT7M5nHsAAAAAFgsinowCgYIKoZIzj0EAwIwgbkx
CzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1FbnRydXN0LCBJbmMuMSgwJgYDVQQLEx9T
ZWUgd3d3LmVudHJ1c3QubmV0L2xlZ2FsLXRlcm1zMTkwNwYDVQQLEzAoYykgMjAx
NiBFbnRydXN0LCBJbmMuIC0gZm9yIGF1dGhvcml6ZWQgdXNlIG9ubHkxLTArBgNV
BAMTJEVudHJ1c3QgZVNJTSBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eTAgFw0xNjEx
MTYxNjA0MDJaGA8yMDUxMTAxNjE2MzQwMlowgbkxCzAJBgNVBAYTAlVTMRYwFAYD
VQQKEw1FbnRydXN0LCBJbmMuMSgwJgYDVQQLEx9TZWUgd3d3LmVudHJ1c3QubmV0
L2xlZ2FsLXRlcm1zMTkwNwYDVQQLEzAoYykgMjAxNiBFbnRydXN0LCBJbmMuIC0g
Zm9yIGF1dGhvcml6ZWQgdXNlIG9ubHkxLTArBgNVBAMTJEVudHJ1c3QgZVNJTSBD
ZXJ0aWZpY2F0aW9uIEF1dGhvcml0eTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IA
BAdzwGHeQ1Wb2f4DmHTByR5/IWL3JugQ1U3908a++bHdlt+TTA7K4c5cYZ+51Yz/
hg/bacxguPDh9uQUK6Wg3a6jcjBwMA8GA1UdEwEB/wQFMAMBAf8wDgYDVR0PAQH/
BAQDAgEGMBcGA1UdIAEB/wQNMAswCQYHZ4ESAQIBADAVBgNVHREEDjAMiApghkgB
hvpsFAoAMB0GA1UdDgQWBBQWcEt/NR42B/GMS3AAXDoAPf1BSjAKBggqhkjOPQQD
AgNJADBGAiEAspjXMvaBZyAg86Z0AAtT0yBRAi1EyaAfNz9kDJeAE04CIQC3efj8
ATL7/tDBOhANy3cK8PS/1NIlu9vqMLCZsZvJ0Q==
-----END CERTIFICATE-----`,
	// MC4 OT ROOT CI v1
	`-----BEGIN CERTIFICATE-----
MIICOjCCAeGgAwIBAgIBATAKBggqhkjOPQQDAjBbMQswCQYDVQQGEwJGUjEeMBwG
A1UEChMVT0JFUlRIVVIgVEVDSE5PTE9HSUVTMRAwDgYDVQQLEwdURUxFQ09NMRow
GAYDVQQDExFNQzQgT1QgUk9PVCBDSSB2MTAeFw0xNjExMTUwMDAwMDFaFw00NjEx
MDgyMzU5NTlaMFsxCzAJBgNVBAYTAkZSMR4wHAYDVQQKExVPQkVSVEhVUiBURUNI
Tk9MT0dJRVMxEDAOBgNVBAsTB1RFTEVDT00xGjAYBgNVBAMTEU1DNCBPVCBST09U
IENJIHYxMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEHb/Gajt3OZxuaDSklBQE
D4lOd6PGPLSvtfkM952ubdyy45tJwAeA0eEii0CLrFT6tcfXkW+H/5mQyMRXaAUk
T6OBlTCBkjAfBgNVHSMEGDAWgBTNbmC3LXoGPLyEYluR6A/jBAbhPjAdBgNVHQ4E
FgQUzW5gty16Bjy8hGJbkegP4wQG4T4wDgYDVR0PAQH/BAQDAgAGMBcGA1UdIAEB
/wQNMAswCQYHZ4ESAQIBADAWBgNVHREEDzANiAsrBgEEAYHvb7OITTAPBgNVHRMB
Af8EBTADAQH/MAoGCCqGSM49BAMCA0cAMEQCIEw4Nc7f2fDtoH+6ON/bknfDQxmT
ikThXjhpLtSrSKN2AiAxHxgC87L0FDnH8dJNlkdGX9c0JIx6oLheIplfS6k+jg==
-----END CERTIFICATE-----`,
}

func (e *Libeuicc) initHttp() {
	e.euiccCtx.http._interface = C.init_http_interface()
}

//export libeuiccHttpTransmit
func libeuiccHttpTransmit(ctx *C.struct_euicc_ctx, url *C.char, rcode *C.uint32_t, rx **C.uint8_t, rx_len *C.uint32_t, tx *C.uint8_t, tx_len C.uint32_t, headers **C.char) C.int {
	r, err := http.NewRequest("POST", C.GoString(url), bytes.NewBuffer(C.GoBytes(unsafe.Pointer(tx), C.int(tx_len))))
	if err != nil {
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

	c := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: rootCAs,
			},
		},
	}

	resp, err := c.Do(r)
	if err != nil {
		return CError
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	*rx = (*C.uint8_t)(C.CBytes(body))
	*rx_len = C.uint32_t(len(body))
	*rcode = C.uint32_t(resp.StatusCode)
	return COK
}
