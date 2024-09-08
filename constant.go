package libeuicc

import "C"

type ProfileState string

type NotificationProfileManagementOperation string

type ProfileIconType string

type ProfilceClass string

type DownloadProgress int

const (
	ProfileStateEnabled  ProfileState = "enabled"
	ProfileStateDisabled ProfileState = "disabled"

	ProfileIconTypeNone ProfileIconType = "none"
	ProfileIconTypeJPG  ProfileIconType = "jpeg"
	ProfileIconTypePNG  ProfileIconType = "png"

	ProfileClassOperational  ProfilceClass = "operational"
	ProfileClassTest         ProfilceClass = "test"
	ProfileClassProvisioning ProfilceClass = "provisioning"

	DownloadProgressGetChallenge             DownloadProgress = 1
	DownloadProgressInitiateAuthentication   DownloadProgress = 2
	DownloadProgressAuthenticateServer       DownloadProgress = 3
	DownloadProgressAuthenticateClient       DownloadProgress = 4
	DownloadProgressConfirmationCodeRequired DownloadProgress = 5
	DownloadProgressConfirmDownload          DownloadProgress = 6
	DownloadProgressPrepareDownload          DownloadProgress = 7
	DownloadProgressGetBoundProfile          DownloadProgress = 9
	DownloadProgressLoadBoundProfile         DownloadProgress = 10

	NotificationProfileManagementOperationDisable NotificationProfileManagementOperation = "disable"
	NotificationProfileManagementOperationEnable  NotificationProfileManagementOperation = "enable"
	NotificationProfileManagementOperationInstall NotificationProfileManagementOperation = "install"
	NotificationProfileManagementOperationDelete  NotificationProfileManagementOperation = "delete"

	CError C.int = -1
	COK    C.int = 0
)
