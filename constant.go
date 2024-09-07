package main

import "C"

type ProfileState string

type NotificationProfileManagementOperation string

const (
	ProfileStateEnabled  ProfileState = "enabled"
	ProfileStateDisabled ProfileState = "disabled"

	NotificationProfileManagementOperationDisable NotificationProfileManagementOperation = "disable"
	NotificationProfileManagementOperationEnable  NotificationProfileManagementOperation = "enable"
	NotificationProfileManagementOperationInstall NotificationProfileManagementOperation = "install"
	NotificationProfileManagementOperationDelete  NotificationProfileManagementOperation = "delete"
)

var CError C.int = -1
