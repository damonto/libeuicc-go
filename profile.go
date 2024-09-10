package libeuicc

/*
#include <stdlib.h>

#include "es10c.h"
#include "tostr.h"
*/
import "C"
import (
	"errors"
	"unsafe"
)

type Profile struct {
	ICCID        string          `json:"iccid"`
	ISDPAid      string          `json:"isdpAid"`
	State        ProfileState    `json:"profileState"`
	Nickname     string          `json:"profileNickname"`
	ProviderName string          `json:"serviceProviderName"`
	ProfileName  string          `json:"profileName"`
	IconType     ProfileIconType `json:"iconType"`
	Icon         string          `json:"icon"`
	Class        ProfilceClass   `json:"profileClass"`
}

func (e *Libeuicc) GetProfiles() ([]*Profile, error) {
	var cProfiles *C.struct_es10c_profile_info_list
	if C.es10c_get_profiles_info(e.ctx, &cProfiles) == CError {
		return nil, errors.New("es10c_get_profiles_info failed")
	}
	defer C.es10c_profile_info_list_free_all(cProfiles)
	profiles := make([]*Profile, 0)
	for cProfile := cProfiles; cProfile != nil; cProfile = cProfile.next {
		profiles = append(profiles, &Profile{
			ICCID:        C.GoString(&cProfile.iccid[0]),
			ISDPAid:      C.GoString(&cProfile.isdpAid[0]),
			State:        ProfileState(C.GoString(C.euicc_profilestate2str(cProfile.profileState))),
			Nickname:     C.GoString(cProfile.profileNickname),
			ProviderName: C.GoString(cProfile.serviceProviderName),
			ProfileName:  C.GoString(cProfile.profileName),
			IconType:     ProfileIconType(C.GoString(C.euicc_icontype2str(cProfile.iconType))),
			Icon:         C.GoString(cProfile.icon),
			Class:        ProfilceClass(C.GoString(C.euicc_profileclass2str(cProfile.profileClass))),
		})
	}
	return profiles, nil
}

func (e *Libeuicc) EnableProfile(iccid string, refrestFlag int) error {
	cIccid := C.CString(iccid)
	defer C.free(unsafe.Pointer(cIccid))
	return e.wrapProfileOperationError(C.es10c_enable_profile(e.ctx, cIccid, C.uint8_t(refrestFlag)))
}

func (e *Libeuicc) DisableProfile(iccid string, refrestFlag int) error {
	cIccid := C.CString(iccid)
	defer C.free(unsafe.Pointer(cIccid))
	return e.wrapProfileOperationError(C.es10c_disable_profile(e.ctx, cIccid, C.uint8_t(refrestFlag)))
}

func (e *Libeuicc) DeleteProfile(iccid string) error {
	cIccid := C.CString(iccid)
	defer C.free(unsafe.Pointer(cIccid))
	return e.wrapProfileOperationError(C.es10c_delete_profile(e.ctx, cIccid))
}

func (e *Libeuicc) wrapProfileOperationError(err C.int) error {
	if err == COK {
		return nil
	}
	switch err {
	case CError:
		return errors.New("failed to operate profile")
	case C.int(1):
		return errors.New("profile not found")
	case C.int(2):
		return errors.New("status conflict")
	case C.int(3):
		return errors.New("the profile is not allowed to do this operation")
	case C.int(4):
		return errors.New("wrong profile reenabling")
	default:
		return errors.New("unknown error")
	}
}

func (e *Libeuicc) SetNickname(iccid, nickname string) error {
	if len(nickname) > 64 {
		return errors.New("nickname is too long")
	}
	cIccid := C.CString(iccid)
	defer C.free(unsafe.Pointer(cIccid))
	if C.es10c_set_nickname(e.ctx, cIccid, C.CString(nickname)) == CError {
		return errors.New("es10c_set_profile_nickname failed")
	}
	return nil
}
