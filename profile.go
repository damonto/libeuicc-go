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
	Iccid        string
	IsdpAid      string
	State        ProfileState
	Nickname     string
	ProviderName string
	ProfileName  string
	IconType     ProfileIconType
	Icon         string
	Class        ProfilceClass
}

// GetProfiles returns the installed profiles.
func (e *Libeuicc) GetProfiles() ([]*Profile, error) {
	var cProfiles *C.struct_es10c_profile_info_list
	if C.es10c_get_profiles_info(e.euiccCtx, &cProfiles) == CError {
		return nil, errors.New("es10c_get_profiles_info failed")
	}
	defer C.es10c_profile_info_list_free_all(cProfiles)
	profiles := make([]*Profile, 0)
	for cProfile := cProfiles; cProfile != nil; cProfile = cProfile.next {
		profiles = append(profiles, &Profile{
			Iccid:        C.GoString(&cProfile.iccid[0]),
			IsdpAid:      C.GoString(&cProfile.isdpAid[0]),
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

// EnableProfile enables the profile with the given ICCID.
// It's recommended to send an enable notification via `ProcessNotification` to the SM-DP+ server.
// Some eUICC chips may require a refresh flag. See the GSMA SGP.22 v2.2.2 for more information.
//
// [SGP.22 v2.2.2]: https://www.gsma.com/solutions-and-impact/technologies/esim/wp-content/uploads/2020/06/SGP.22-v2.2.2.pdf#page=82
func (e *Libeuicc) EnableProfile(iccid string, refresh bool) error {
	cIccid := C.CString(iccid)
	defer C.free(unsafe.Pointer(cIccid))
	refreshFlag := C.uint8_t(0)
	if refresh {
		refreshFlag = C.uint8_t(1)
	}
	return e.wrapProfileOperationError(C.es10c_enable_profile(e.euiccCtx, cIccid, refreshFlag))
}

// DisableProfile disables the profile with the given ICCID.
// It's recommanded to send a disable notification via `ProcessNotification` to the SM-DP+ server.
// Some eUICC chips may require a refresh flag. See the GSMA SGP.22 v2.2.2 for more information.
//
// [SGP.22 v2.2.2]: https://www.gsma.com/solutions-and-impact/technologies/esim/wp-content/uploads/2020/06/SGP.22-v2.2.2.pdf#page=86
func (e *Libeuicc) DisableProfile(iccid string, refresh bool) error {
	cIccid := C.CString(iccid)
	defer C.free(unsafe.Pointer(cIccid))
	refreshFlag := C.uint8_t(0)
	if refresh {
		refreshFlag = C.uint8_t(1)
	}
	return e.wrapProfileOperationError(C.es10c_disable_profile(e.euiccCtx, cIccid, refreshFlag))
}

// DeleteProfile deletes the profile with the given ICCID.
// Once a profile is deleted, it cannot be recovered, and you must send the delete notification via `ProcessNotification` to the SM-DP+ server. otherwise, you may can't install it on another one of your devices.
func (e *Libeuicc) DeleteProfile(iccid string) error {
	cIccid := C.CString(iccid)
	defer C.free(unsafe.Pointer(cIccid))
	return e.wrapProfileOperationError(C.es10c_delete_profile(e.euiccCtx, cIccid))
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

// SetNickname sets the nickname of the profile with the given ICCID. The nickname cannot be longer than 64 UTF-8 bytes.
func (e *Libeuicc) SetNickname(iccid, nickname string) error {
	if len(nickname) > 64 {
		return errors.New("nickname is too long")
	}
	cIccid := C.CString(iccid)
	defer C.free(unsafe.Pointer(cIccid))
	if C.es10c_set_nickname(e.euiccCtx, cIccid, C.CString(nickname)) == CError {
		return errors.New("es10c_set_profile_nickname failed")
	}
	return nil
}
