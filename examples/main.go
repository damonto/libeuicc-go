package main

import (
	"github.com/damonto/libeuicc-go"
)

func main() {
	euicc, err := libeuicc.NewLibeuicc(nil)
	if err != nil {
		panic(err)
	}
	defer euicc.Free()
}
