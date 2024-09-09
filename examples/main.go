package main

import (
	"fmt"

	"github.com/damonto/libeuicc-go"
)

func main() {
	pcscReader, err := libeuicc.NewPCSCReader()
	if err != nil {
		fmt.Println(err)
		return
	}
	euicc, err := libeuicc.NewLibeuicc(pcscReader)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer euicc.Free()

	fmt.Println(euicc.GetEid())
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	// err = euicc.DownloadProfile(ctx, &libeuicc.ActivationCode{
	// 	SMDP:       "millicomelsalvador.validereachdpplus.com",
	// 	MatchingId: "GENERICJOWMI-FAHTCU0-SKFMYPW6UIEFGRWC8GE933ITFAUVN63WMUVHFOWTS80",
	// }, &libeuicc.DownloadOption{
	// 	ProgressBar: func(progress libeuicc.DownloadProgress) {
	// 		fmt.Println(progress)
	// 	},
	// 	ConfirmFunc: func(metadata *libeuicc.ProfileMetadata) bool {
	// 		fmt.Println(metadata)
	// 		return false
	// 	},
	// 	ConfirmationCodeFunc: func() string {
	// 		return ""
	// 	},
	// })

	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println("Download profile success")

	// sig := make(chan os.Signal, 1)
	// signal.Notify(sig, os.Interrupt)
	// <-sig
	// cancel()
}
