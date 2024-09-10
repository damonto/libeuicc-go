package libeuicc

/*
#include <stdlib.h>

#include "es10a.h"
#include "es10c.h"
#include "es10c_ex.h"
*/
import "C"
import (
	"errors"
	"unsafe"
)

type EuiccInfo2 struct {
	SasAccreditationNumber string          `json:"sasAcreditationNumber"`
	ProfileVersion         string          `json:"profileVersion"`
	FirmwareVersion        string          `json:"firmwareVersion"`
	ExtCardResource        ExtCardResource `json:"extCardResource"`
	CiPKIdForSigning       []string        `json:"ciPKIdListForSigning"`
}

type ConfiguredAddresses struct {
	DefaultDPAddress string `json:"defaultDpAddress"`
	RootDSAddress    string `json:"rootDsAddress"`
}

type ExtCardResource struct {
	FreeNonVolatileMemory int `json:"freeNonVolatileMemory"`
	FreeVolatileMemory    int `json:"freeVolatileMemory"`
}

func (e *Libeuicc) GetEid() (string, error) {
	var eid *C.char
	if C.es10c_get_eid(e.ctx, &eid) == CError {
		return "", errors.New("es10c_get_eid failed")
	}
	defer C.free(unsafe.Pointer(eid))
	return C.GoString(eid), nil
}

func (e *Libeuicc) GetEuiccInfo2() (*EuiccInfo2, error) {
	var euiccInfo2 C.struct_es10c_ex_euiccinfo2
	if C.es10c_ex_get_euiccinfo2(e.ctx, &euiccInfo2) == CError {
		return nil, errors.New("es10c_ex_get_euiccinfo2 failed")
	}
	defer C.es10c_ex_euiccinfo2_free(&euiccInfo2)
	return &EuiccInfo2{
		SasAccreditationNumber: C.GoString(euiccInfo2.sasAcreditationNumber),
		ProfileVersion:         C.GoString(euiccInfo2.profileVersion),
		FirmwareVersion:        C.GoString(euiccInfo2.euiccFirmwareVer),
		ExtCardResource: ExtCardResource{
			FreeNonVolatileMemory: int(euiccInfo2.extCardResource.freeNonVolatileMemory),
			FreeVolatileMemory:    int(euiccInfo2.extCardResource.freeVolatileMemory),
		},
		CiPKIdForSigning: GoStrings(euiccInfo2.euiccCiPKIdListForSigning),
	}, nil
}

func (e *Libeuicc) GetConfiguredAddresses() (*ConfiguredAddresses, error) {
	var configuredAddresses C.struct_es10a_euicc_configured_addresses
	if C.es10a_get_euicc_configured_addresses(e.ctx, &configuredAddresses) == CError {
		return nil, errors.New("es10a_get_euicc_configured_addresses failed")
	}
	defer C.es10a_euicc_configured_addresses_free(&configuredAddresses)
	return &ConfiguredAddresses{
		DefaultDPAddress: C.GoString(configuredAddresses.defaultDpAddress),
		RootDSAddress:    C.GoString(configuredAddresses.rootDsAddress),
	}, nil
}

func (e *Libeuicc) Reset() error {
	if C.es10c_euicc_memory_reset(e.ctx) == CError {
		return errors.New("es10c_euicc_memory_reset failed")
	}
	return nil
}

func (e *Libeuicc) SetDefaultSMDPAddress(address string) error {
	cAddress := C.CString(address)
	defer C.free(unsafe.Pointer(cAddress))
	if C.es10a_set_default_dp_address(e.ctx, cAddress) == CError {
		return errors.New("es10c_set_default_smdp_address failed")
	}
	return nil
}
