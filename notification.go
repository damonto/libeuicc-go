package main

/*
#cgo CFLAGS: -I${SRCDIR}/lpac/cjson

#include "es9p.h"
#include "es10b.h"
#include "tostr.h"
*/
import "C"
import (
	"errors"
	"unsafe"
)

type Notification struct {
	SeqNumber                  int    `json:"seqNumber"`
	ProfileManagementOperation string `json:"profileManagementOperation"`
	NotificationAddress        string `json:"notificationAddress"`
	ICCID                      string `json:"iccid"`
}

const (
	NotificationProfileManagementOperationDisable = "disable"
	NotificationProfileManagementOperationEnable  = "enable"
	NotificationProfileManagementOperationInstall = "install"
	NotificationProfileManagementOperationDelete  = "delete"
)

func (e *Libeuicc) GetNotifications() ([]*Notification, error) {
	notifications := (**C.struct_es10b_notification_metadata_list)(C.malloc(C.sizeof_struct_es10b_notification_metadata_list))
	if C.es10b_list_notification(e.euiccCtx, notifications) == CError {
		return nil, errors.New("es10b_list_notification failed")
	}
	result := make([]*Notification, 0)
	for i := 0; ; i++ {
		notification := *(**C.struct_es10b_notification_metadata_list)(unsafe.Pointer(uintptr(unsafe.Pointer(notifications)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))
		if notification == nil {
			break
		}
		result = append(result, &Notification{
			SeqNumber:                  int(notification.seqNumber),
			ProfileManagementOperation: C.GoString(C.euicc_profilemanagementoperation2str(notification.profileManagementOperation)),
			NotificationAddress:        C.GoString(notification.notificationAddress),
			ICCID:                      C.GoString(notification.iccid),
		})
	}
	return result, nil
}

func (e *Libeuicc) ProcessNotification(seqNumber int, remove bool) error {
	var notification C.struct_es10b_pending_notification
	if C.es10b_retrieve_notifications_list(e.euiccCtx, &notification, C.ulong(seqNumber)) == CError {
		return errors.New("es10b_retrieve_notifications_list failed")
	}
	e.euiccCtx.http.server_address = notification.notificationAddress
	if C.es9p_handle_notification(e.euiccCtx, notification.b64_PendingNotification) == CError {
		return errors.New("es9p_handle_notification failed")
	}
	if remove {
		if C.es10b_remove_notification_from_list(e.euiccCtx, C.ulong(seqNumber)) == CError {
			return errors.New("es10b_remove_notification failed")
		}
	}
	return nil
}

func (e *Libeuicc) DeleteNotification(seqNumber int) error {
	if C.es10b_remove_notification_from_list(e.euiccCtx, C.ulong(seqNumber)) == CError {
		return errors.New("es10b_remove_notification failed")
	}
	return nil
}
