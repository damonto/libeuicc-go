package libeuicc

/*
#include <stdlib.h>
#include <string.h>

#include "es10a.h"
#include "es10b.h"
#include "es9p.h"
#include "es8p.h"
#include "tostr.h"
#include "derutil.h"
*/
import "C"
import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"unsafe"
)

type ActivationCode struct {
	// SMDP is the address of the SM-DP+ server.
	SMDP string
	// MatchingId is the matching ID. (Optional)
	MatchingId string
	// ConfirmationCode is the confirmation code. (Optional)
	ConfirmationCode string
	// IMEI is the IMEI of the device. (Optional)
	IMEI string
}

type ProfileMetadata struct {
	Iccid        string
	Nickname     string
	ProviderName string
	ProfileName  string
	IconType     ProfileIconType
	Icon         string
}

type DownloadOption struct {
	// ProgressBar is the callback function to handle the download progress.
	ProgressBar func(progress DownloadProgress)

	// ConfirmFunc is the callback function to confirm the download.
	// If it is not set, the download will be confirmed automatically.
	// If the function returns false, the download will be canceled.
	ConfirmFunc func(metadata *ProfileMetadata) bool

	// ConfirmationCodeFunc is the callback function to get the confirmation code.
	// If you do not set it and a confirmation code is required, the download will be canceled.
	ConfirmationCodeFunc func() string
}

var (
	ErrDownloadCanceled         = errors.New("download canceled")
	ErrConfirmationCodeRequired = errors.New("confirmation code required")
)

// DownloadProfile downloads the eSIM profile from the SM-DP+ server.
func (e *Libeuicc) DownloadProfile(ctx context.Context, activationCode *ActivationCode, downloadOption *DownloadOption) error {
	defer e.cleanupHttp()

	cSmdp := C.CString(activationCode.SMDP)
	cMatchingId := C.CString(activationCode.MatchingId)
	cImei := C.CString(activationCode.IMEI)
	cConfirmationCode := C.CString(activationCode.ConfirmationCode)
	defer C.free(unsafe.Pointer(cMatchingId))
	defer C.free(unsafe.Pointer(cImei))
	defer C.free(unsafe.Pointer(cSmdp))
	defer C.free(unsafe.Pointer(cConfirmationCode))

	logger.Debug("Downloading profile", "smdp", activationCode.SMDP, "matchingId", activationCode.MatchingId, "imei", activationCode.IMEI, "confirmationCode", activationCode.ConfirmationCode)

	e.euiccCtx.http.server_address = cSmdp

	e.handleProgress(downloadOption, DownloadProgressGetChallenge)
	if (C.es10b_get_euicc_challenge_and_info(e.euiccCtx)) == CError {
		return errors.New("es10b_get_euicc_challenge_and_info failed")
	}

	e.handleProgress(downloadOption, DownloadProgressInitiateAuthentication)
	if C.es9p_initiate_authentication(e.euiccCtx) == CError {
		return errors.New("es9p_initiate_authentication failed")
	}

	e.handleProgress(downloadOption, DownloadProgressAuthenticateServer)
	if C.es10b_authenticate_server(e.euiccCtx, cMatchingId, cImei) == CError {
		return errors.New("es10b_authenticate_server failed")
	}

	e.handleProgress(downloadOption, DownloadProgressAuthenticateClient)
	if C.es9p_authenticate_client(e.euiccCtx) != COK {
		return e.wrapES9PError()
	}

	ccRequired, err := e.isConfirmationCodeRequired()
	if err != nil {
		if cancelErr := e.cancelSession(); cancelErr != nil {
			return errors.Join(err, cancelErr)
		}
		return err
	}
	if ccRequired && activationCode.ConfirmationCode == "" {
		e.handleProgress(downloadOption, DownloadProgressConfirmationCodeRequired)
		if downloadOption != nil && downloadOption.ConfirmationCodeFunc != nil {
			cConfirmationCode = C.CString(downloadOption.ConfirmationCodeFunc())
		}
		if cConfirmationCode == nil {
			if err := e.cancelSession(); err != nil {
				return errors.Join(ErrConfirmationCodeRequired, err)
			}
			return ErrConfirmationCodeRequired
		}
	}

	profileMetadata, err := e.parseProfileMetadata()
	if err != nil {
		if cancelErr := e.cancelSession(); cancelErr != nil {
			return errors.Join(err, cancelErr)
		}
		return err
	}
	e.handleProgress(downloadOption, DownloadProgressConfirmDownload)
	if downloadOption != nil && downloadOption.ConfirmFunc != nil {
		confirmDownload := downloadOption.ConfirmFunc(profileMetadata)
		if !confirmDownload {
			if err := e.cancelSession(); err != nil {
				return errors.Join(ErrDownloadCanceled, err)
			}
			return ErrDownloadCanceled
		}
	}

	if e.isCanceled(ctx) {
		return e.cancelSession()
	}
	e.handleProgress(downloadOption, DownloadProgressPrepareDownload)
	if C.es10b_prepare_download(e.euiccCtx, cConfirmationCode) == CError {
		return errors.New("es10b_prepare_download failed")
	}

	if e.isCanceled(ctx) {
		return e.cancelSession()
	}
	e.handleProgress(downloadOption, DownloadProgressGetBoundProfile)
	if C.es9p_get_bound_profile_package(e.euiccCtx) == CError {
		return errors.New("es9p_get_bound_profile_package failed")
	}

	if e.isCanceled(ctx) {
		return e.cancelSession()
	}
	downloadResult := (*C.struct_es10b_load_bound_profile_package_result)(C.malloc(C.sizeof_struct_es10b_load_bound_profile_package_result))
	if downloadResult == nil {
		if err := e.cancelSession(); err != nil {
			return errors.Join(errors.New("failed to allocate memory for downloadResult"), err)
		}
		return errors.New("failed to allocate memory for downloadResult")
	}
	C.memset(unsafe.Pointer(downloadResult), 0, C.sizeof_struct_es10b_load_bound_profile_package_result)
	defer C.free(unsafe.Pointer(downloadResult))
	e.handleProgress(downloadOption, DownloadProgressLoadBoundProfile)
	if C.es10b_load_bound_profile_package(e.euiccCtx, downloadResult) != COK {
		if err := e.cancelSession(); err != nil {
			return errors.Join(e.wrapLoadBPPError(downloadResult), err)
		}
		return e.wrapLoadBPPError(downloadResult)
	}
	return nil
}

func (e *Libeuicc) handleProgress(downloadOption *DownloadOption, progress DownloadProgress) {
	if downloadOption == nil || downloadOption.ProgressBar == nil {
		return
	}
	downloadOption.ProgressBar(progress)
}

func (e *Libeuicc) wrapLoadBPPError(downloadResult *C.struct_es10b_load_bound_profile_package_result) error {
	return fmt.Errorf("bppCommandId: %d bppCommand: %s errorReasonId: %d errorReason: %s",
		downloadResult.bppCommandId,
		C.GoString(C.euicc_bppcommandid2str(downloadResult.bppCommandId)),
		downloadResult.errorReason,
		C.GoString(C.euicc_errorreason2str(downloadResult.errorReason)),
	)
}

func (e *Libeuicc) wrapES9PError() error {
	return fmt.Errorf("subjectIdentifier: %s subjectCode: %s reasonCode: %s message: %s",
		C.GoString(&e.euiccCtx.http.status.subjectIdentifier[0]),
		C.GoString(&e.euiccCtx.http.status.subjectCode[0]),
		C.GoString(&e.euiccCtx.http.status.reasonCode[0]),
		C.GoString(&e.euiccCtx.http.status.message[0]),
	)
}

func (e *Libeuicc) isCanceled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func (e *Libeuicc) cancelSession() error {
	if C.es10b_cancel_session(e.euiccCtx, C.ES10B_CANCEL_SESSION_REASON_ENDUSERREJECTION) == CError {
		return errors.New("es10b_cancel_session failed")
	}
	if C.es9p_cancel_session(e.euiccCtx) == CError {
		return errors.New("es9p_cancel_session failed")
	}
	return nil
}

func (e *Libeuicc) isConfirmationCodeRequired() (bool, error) {
	base64decodedSmdpSigned2, err := base64.StdEncoding.DecodeString(C.GoString(e.euiccCtx.http._internal.prepare_download_param.b64_smdpSigned2))
	if err != nil {
		return false, err
	}
	cB64s := C.CString(string(base64decodedSmdpSigned2))
	defer C.free(unsafe.Pointer(cB64s))

	smdpSigned2 := (*C.struct_euicc_derutil_node)(C.malloc(C.sizeof_struct_euicc_derutil_node))
	if smdpSigned2 == nil {
		return false, errors.New("failed to allocate memory for ccFlag")
	}
	C.memset(unsafe.Pointer(smdpSigned2), 0, C.sizeof_struct_euicc_derutil_node)
	defer C.free(unsafe.Pointer(smdpSigned2))
	if C.euicc_derutil_unpack_find_tag(smdpSigned2, 0x30, (*C.uchar)(unsafe.Pointer(cB64s)), C.uint(len(base64decodedSmdpSigned2))) == CError {
		return false, errors.New("euicc_derutil_unpack_find_tag failed")
	}

	ccFlag := (*C.struct_euicc_derutil_node)(C.malloc(C.sizeof_struct_euicc_derutil_node))
	if ccFlag == nil {
		return false, errors.New("failed to allocate memory for ccFlag")
	}
	C.memset(unsafe.Pointer(ccFlag), 0, C.sizeof_struct_euicc_derutil_node)
	defer C.free(unsafe.Pointer(ccFlag))
	if C.euicc_derutil_unpack_find_tag(ccFlag, 0x01, smdpSigned2.value, smdpSigned2.length) == CError {
		return false, errors.New("euicc_derutil_unpack_find_tag failed")
	}
	return C.euicc_derutil_convert_bin2long(ccFlag.value, ccFlag.length) != 0, nil
}

func (e *Libeuicc) parseProfileMetadata() (*ProfileMetadata, error) {
	var cProfileMetadata *C.struct_es8p_metadata
	if C.es8p_metadata_parse(&cProfileMetadata, e.euiccCtx.http._internal.prepare_download_param.b64_profileMetadata) == CError {
		return nil, errors.New("es8p_parse_metadata failed")
	}
	defer C.es8p_metadata_free(&cProfileMetadata)
	return &ProfileMetadata{
		Iccid:        C.GoString(&cProfileMetadata.iccid[0]),
		ProviderName: C.GoString(cProfileMetadata.serviceProviderName),
		ProfileName:  C.GoString(cProfileMetadata.profileName),
		IconType:     ProfileIconType(C.GoString(C.euicc_icontype2str(cProfileMetadata.iconType))),
		Icon:         C.GoString(cProfileMetadata.icon),
	}, nil
}
