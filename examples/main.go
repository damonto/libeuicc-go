package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	err = euicc.DownloadProfile(ctx, &libeuicc.ActivationCode{
		SMDP:       "rsp.septs.app",
		MatchingId: "123413231",
	}, &libeuicc.DownloadOption{
		ProgressBar: func(progress libeuicc.DownloadProgress) {
			fmt.Println(progress)
		},
		ConfirmationCodeFunc: func() string {
			fmt.Println("Please input confirmation code:")
			return ""
		},
		ConfirmFunc: func(metadata *libeuicc.ProfileMetadata) bool {
			fmt.Println(metadata)
			return false
		},
	})

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Download profile success")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	cancel()
}
