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
	SasAccreditationNumber string
	ProfileVersion         string
	FirmwareVersion        string
	ExtCardResource        ExtCardResource
	CiPKIdForSigning       []string
}

type ConfiguredAddresses struct {
	DefaultDPAddress string
	RootDSAddress    string
}

type ExtCardResource struct {
	FreeNonVolatileMemory int
	FreeVolatileMemory    int
}

// GetEid returns the EID of the eUICC.
func (e *Libeuicc) GetEid() (string, error) {
	var eid *C.char
	if C.es10c_get_eid(e.euiccCtx, &eid) == CError {
		return "", errors.New("es10c_get_eid failed")
	}
	defer C.free(unsafe.Pointer(eid))
	return C.GoString(eid), nil
}

// GetEuiccInfo2 returns the eUICC information.
// It includes the SAS accreditation number, profile version, firmware version, card resource, and CI PKI ID list for signing.
func (e *Libeuicc) GetEuiccInfo2() (*EuiccInfo2, error) {
	var euiccInfo2 C.struct_es10c_ex_euiccinfo2
	if C.es10c_ex_get_euiccinfo2(e.euiccCtx, &euiccInfo2) == CError {
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

// GetConfiguredAddresses returns the configured addresses of the eUICC.
// It includes the default SM-DP+ address and root SM-DS address.
func (e *Libeuicc) GetConfiguredAddresses() (*ConfiguredAddresses, error) {
	var configuredAddresses C.struct_es10a_euicc_configured_addresses
	if C.es10a_get_euicc_configured_addresses(e.euiccCtx, &configuredAddresses) == CError {
		return nil, errors.New("es10a_get_euicc_configured_addresses failed")
	}
	defer C.es10a_euicc_configured_addresses_free(&configuredAddresses)
	return &ConfiguredAddresses{
		DefaultDPAddress: C.GoString(configuredAddresses.defaultDpAddress),
		RootDSAddress:    C.GoString(configuredAddresses.rootDsAddress),
	}, nil
}

// Reset resets the eUICC memory.
// Attention: This operation will erase all the data on the eUICC. 
func (e *Libeuicc) Reset() error {
	if C.es10c_euicc_memory_reset(e.euiccCtx) == CError {
		return errors.New("es10c_euicc_memory_reset failed")
	}
	return nil
}

// SetDefaultSMDPAddress sets the default SM-DP+ address of the eUICC.
func (e *Libeuicc) SetDefaultSMDPAddress(address string) error {
	cAddress := C.CString(address)
	defer C.free(unsafe.Pointer(cAddress))
	if C.es10a_set_default_dp_address(e.euiccCtx, cAddress) == CError {
		return errors.New("es10c_set_default_smdp_address failed")
	}
	return nil
}
