package libeuicc

/*
#include "es9p.h"
#include "es10b.h"
#include "tostr.h"
*/
import "C"
import (
	"errors"
)

type Notification struct {
	SeqNumber                  int
	ProfileManagementOperation NotificationProfileManagementOperation
	NotificationAddress        string
	Iccid                      string
}

// GetNotifications returns the list of notifications.
func (e *Libeuicc) GetNotifications() ([]*Notification, error) {
	var cNotifications *C.struct_es10b_notification_metadata_list
	if C.es10b_list_notification(e.euiccCtx, &cNotifications) == CError {
		return nil, errors.New("es10b_list_notification failed")
	}
	defer C.es10b_notification_metadata_list_free_all(cNotifications)
	var notifications []*Notification
	for cNotification := cNotifications; cNotification != nil; cNotification = cNotification.next {
		notifications = append(notifications, &Notification{
			SeqNumber:                  int(cNotification.seqNumber),
			ProfileManagementOperation: NotificationProfileManagementOperation(C.GoString(C.euicc_profilemanagementoperation2str(cNotification.profileManagementOperation))),
			NotificationAddress:        C.GoString(cNotification.notificationAddress),
			Iccid:                      C.GoString(cNotification.iccid),
		})
	}
	return notifications, nil
}

// ProcessNotification processes the notification with the given sequence number.
// If remove is true, the notification will be removed from the eUICC.
func (e *Libeuicc) ProcessNotification(seqNumber int, remove bool) error {
	defer e.cleanupHttp()
	var notification C.struct_es10b_pending_notification
	if C.es10b_retrieve_notifications_list(e.euiccCtx, &notification, C.ulong(seqNumber)) == CError {
		return errors.New("es10b_retrieve_notifications_list failed")
	}
	e.euiccCtx.http.server_address = notification.notificationAddress
	if C.es9p_handle_notification(e.euiccCtx, notification.b64_PendingNotification) == CError {
		return errors.New("es9p_handle_notification failed")
	}
	defer C.es10b_pending_notification_free(&notification)
	if remove {
		if C.es10b_remove_notification_from_list(e.euiccCtx, C.ulong(seqNumber)) == CError {
			return errors.New("es10b_remove_notification failed")
		}
	}
	return nil
}

// DeleteNotification deletes the notification with the given sequence number.
func (e *Libeuicc) DeleteNotification(seqNumber int) error {
	if C.es10b_remove_notification_from_list(e.euiccCtx, C.ulong(seqNumber)) == CError {
		return errors.New("es10b_remove_notification failed")
	}
	return nil
}
