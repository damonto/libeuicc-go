package main

type Profile struct {
	ICCID        string       `json:"iccid"`
	ISDPAid      string       `json:"isdpAid"`
	State        ProfileState `json:"profileState"`
	Nickname     string       `json:"profileNickname"`
	ProviderName string       `json:"serviceProviderName"`
	ProfileName  string       `json:"profileName"`
	IconType     string       `json:"iconType"`
	Icon         string       `json:"icon"`
	Class        string       `json:"profileClass"`
}
