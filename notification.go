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
	SeqNumber                  int                                    `json:"seqNumber"`
	ProfileManagementOperation NotificationProfileManagementOperation `json:"profileManagementOperation"`
	NotificationAddress        string                                 `json:"notificationAddress"`
	ICCID                      string                                 `json:"iccid"`
}

func (e *Libeuicc) GetNotifications() ([]*Notification, error) {
	var cNotifications *C.struct_es10b_notification_metadata_list
	if C.es10b_list_notification(e.ctx, &cNotifications) == CError {
		return nil, errors.New("es10b_list_notification failed")
	}
	defer C.es10b_notification_metadata_list_free_all(cNotifications)
	notifications := make([]*Notification, 0)
	for cNotification := cNotifications; cNotification != nil; cNotification = cNotification.next {
		notifications = append(notifications, &Notification{
			SeqNumber:                  int(cNotification.seqNumber),
			ProfileManagementOperation: NotificationProfileManagementOperation(C.GoString(C.euicc_profilemanagementoperation2str(cNotification.profileManagementOperation))),
			NotificationAddress:        C.GoString(cNotification.notificationAddress),
			ICCID:                      C.GoString(cNotification.iccid),
		})
	}
	return notifications, nil
}

func (e *Libeuicc) ProcessNotification(seqNumber int, remove bool) error {
	defer e.cleanupHttp()
	var notification C.struct_es10b_pending_notification
	if C.es10b_retrieve_notifications_list(e.ctx, &notification, C.ulong(seqNumber)) == CError {
		return errors.New("es10b_retrieve_notifications_list failed")
	}
	e.ctx.http.server_address = notification.notificationAddress
	if C.es9p_handle_notification(e.ctx, notification.b64_PendingNotification) == CError {
		return errors.New("es9p_handle_notification failed")
	}
	defer C.es10b_pending_notification_free(&notification)
	if remove {
		if C.es10b_remove_notification_from_list(e.ctx, C.ulong(seqNumber)) == CError {
			return errors.New("es10b_remove_notification failed")
		}
	}
	return nil
}

func (e *Libeuicc) DeleteNotification(seqNumber int) error {
	if C.es10b_remove_notification_from_list(e.ctx, C.ulong(seqNumber)) == CError {
		return errors.New("es10b_remove_notification failed")
	}
	return nil
}
