package libeuicc

/*
#include <stdlib.h>

#include "es10b.h"
#include "es9p.h"
*/
import "C"
import (
	"errors"
	"unsafe"
)

type RspServerAddress struct {
	RspServerAddress string `json:"rspServerAddress"`
}

var defaultSmds = []string{
	"lpa.live.esimdiscovery.com",
	"lpa.ds.gsma.com",
}

// Discover discovers the RSP server address.
// If smds is empty, it will try to discover the RSP server address from the default SM-DS servers.
func (e *Libeuicc) Discover(smds string, imei string) ([]*RspServerAddress, error) {
	var discoveryResult []*RspServerAddress
	if smds == "" {
		for _, s := range defaultSmds {
			r, err := e.discover(s, imei)
			if err != nil {
				return nil, err
			}
			discoveryResult = append(discoveryResult, r...)
		}
		return discoveryResult, nil
	}

	r, err := e.discover(smds, imei)
	if err != nil {
		return nil, err
	}
	discoveryResult = append(discoveryResult, r...)
	return discoveryResult, nil
}

func (e *Libeuicc) discover(smds string, imei string) ([]*RspServerAddress, error) {
	defer e.cleanupHttp()

	cImei := C.CString(imei)
	cSmds := C.CString(smds)
	defer C.free(unsafe.Pointer(cSmds))
	defer C.free(unsafe.Pointer(cImei))

	e.euiccCtx.http.server_address = cSmds
	if C.es10b_get_euicc_challenge_and_info(e.euiccCtx) == CError {
		return nil, errors.New("es10b_get_euicc_challenge_and_info failed")
	}
	if C.es9p_initiate_authentication(e.euiccCtx) == CError {
		return nil, errors.New("es9p_initiate_authentication failed")
	}
	if C.es10b_authenticate_server(e.euiccCtx, cImei, nil) == CError {
		return nil, errors.New("es10b_authenticate_server failed")
	}

	var cSmdpAdresses **C.char
	if C.es11_authenticate_client(e.euiccCtx, &cSmdpAdresses) == CError {
		return nil, errors.New("es11_authenticate_client failed")
	}
	defer C.es11_smdp_list_free_all(cSmdpAdresses)

	rspServerAddresses := make([]*RspServerAddress, 0)
	for _, smdpAddress := range GoStrings(cSmdpAdresses) {
		rspServerAddresses = append(rspServerAddresses, &RspServerAddress{
			RspServerAddress: smdpAddress,
		})
	}
	return rspServerAddresses, nil
}
