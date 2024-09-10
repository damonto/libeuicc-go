# Libeuicc Go bindings

### Example
```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/damonto/libeuicc-go"
	"github.com/damonto/libeuicc-go/driver/pcsc"
)

func main() {
	pcscReader, err := pcsc.NewPCSCReader()
	if err != nil {
		fmt.Println(err)
		return
	}
	euicc, err := libeuicc.NewLibeuicc(pcscReader, libeuicc.NewDefaultLogger(libeuicc.LogErrorLevel))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer euicc.Free()

	fmt.Println(euicc.GetEid())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	err = euicc.DownloadProfile(ctx, &libeuicc.ActivationCode{
		SMDP:       "rsp.septs.app",
		MatchingId: "123413231",
	}, &libeuicc.DownloadOption{
		ProgressBar: func(progress libeuicc.DownloadProgress) {
			// If you want to download, please remove this line.
			if progress == libeuicc.DownloadProgressGetBoundProfile {
				cancel()
			}
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
}
```
