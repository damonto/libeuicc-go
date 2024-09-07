package main

type ActivationCode struct {
	SMDP             string
	MatchingId       string
	ConfirmationCode string
	IMEI             string
}

type ProfileMetadata struct {
	ICCID        string `json:"iccid"`
	Nickname     string `json:"profileNickname"`
	ProviderName string `json:"serviceProviderName"`
	ProfileName  string `json:"profileName"`
	IconType     string `json:"iconType"`
	Icon         string `json:"icon"`
}
