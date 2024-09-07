package main

/*
#include "es10a.h"
#include "es10c.h"
#include "es10c_ex.h"
*/
import "C"
import (
	"errors"
)

type EuiccInfo2 struct {
	SasAccreditationNumber string          `json:"sasAcreditationNumber"`
	ProfileVersion         string          `json:"profileVersion"`
	FirmwareVersion        string          `json:"firmwareVersion"`
	ExtCardResource        ExtCardResource `json:"extCardResource"`
	PkiForSigning          []string        `json:"ciPKIdListForSigning"`
}

type ConfiguredAddresses struct {
	DefaultDPAddress string `json:"defaultDpAddress"`
	RootDSAddress    string `json:"rootDsAddress"`
}

type ExtCardResource struct {
	FreeNonVolatileMemory int `json:"freeNonVolatileMemory"`
	FreeVolatileMemory    int `json:"freeVolatileMemory"`
}

func (e *Libeuicc) GetEid() string {
	var eid *C.char
	C.es10c_get_eid(e.euiccCtx, &eid)
	return C.GoString(eid)
}

func (e *Libeuicc) GetEuiccInfo2() (*EuiccInfo2, error) {
	var euiccInfo2 C.struct_es10c_ex_euiccinfo2
	if C.es10c_ex_get_euiccinfo2(e.euiccCtx, &euiccInfo2) == CError {
		return nil, errors.New("es10c_ex_get_euiccinfo2 failed")
	}
	return &EuiccInfo2{
		SasAccreditationNumber: C.GoString(euiccInfo2.sasAcreditationNumber),
		ProfileVersion:         C.GoString(euiccInfo2.profileVersion),
		FirmwareVersion:        C.GoString(euiccInfo2.euiccFirmwareVer),
		ExtCardResource: ExtCardResource{
			FreeNonVolatileMemory: int(euiccInfo2.extCardResource.freeNonVolatileMemory),
			FreeVolatileMemory:    int(euiccInfo2.extCardResource.freeVolatileMemory),
		},
		PkiForSigning: GoStrings(euiccInfo2.euiccCiPKIdListForSigning),
	}, nil
}

func (e *Libeuicc) GetConfiguredAddresses() (*ConfiguredAddresses, error) {
	var configuredAddresses C.struct_es10a_euicc_configured_addresses
	if C.es10a_get_euicc_configured_addresses(e.euiccCtx, &configuredAddresses) == CError {
		return nil, errors.New("es10a_get_euicc_configured_addresses failed")
	}
	return &ConfiguredAddresses{
		DefaultDPAddress: C.GoString(configuredAddresses.defaultDpAddress),
		RootDSAddress:    C.GoString(configuredAddresses.rootDsAddress),
	}, nil
}

func (e Libeuicc) Purge() error {
	if C.es10c_euicc_memory_reset(e.euiccCtx) == CError {
		return errors.New("es10c_euicc_memory_reset failed")
	}
	return nil
}
