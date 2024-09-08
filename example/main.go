package main

import (
	"fmt"

	"github.com/damonto/libeuicc-go"
)

func main() {
	pcscReader, err := libeuicc.NewPCSCReader()
	if err != nil {
		return
	}
	euicc, err := libeuicc.NewLibeuicc(pcscReader)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer euicc.Free()
	fmt.Println(euicc.GetEid())
	profiles, err := euicc.GetProfiles()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, profile := range profiles {
		fmt.Println(profile.ProfileName, profile.ICCID)
	}
}
