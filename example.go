package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

type echoAPDU struct {
	conn    net.Conn
	channel byte
}

func NewEchoApdu() APDU {
	conn, err := net.Dial("tcp", "127.0.0.1:9527")
	if err != nil {
		log.Println("ERROR: Dial failed", err)
		panic(err)
	}

	return &echoAPDU{
		conn: conn,
	}
}

func (a *echoAPDU) Connect() error {
	return nil
}

func (a *echoAPDU) Disconnect() error {
	return a.conn.Close()
}

func (a *echoAPDU) transmit(command []byte) ([]byte, error) {
	if _, err := a.conn.Write(command); err != nil {
		return nil, err
	}
	buf := make([]byte, 1024)
	n, err := a.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (a *echoAPDU) Transmit(command []byte) ([]byte, error) {
	resp, err := a.transmit(command)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (a *echoAPDU) OpenLogicalChannel(aid []byte) (int, error) {
	logicalChannelResp, err := a.transmit([]byte{0x00, 0x70, 0x00, 0x00, 0x01})
	if err != nil {
		return 0, err
	}
	a.channel = logicalChannelResp[0]
	selectCmd := []byte{a.channel, 0xA4, 0x04, 0x00, byte(len(aid))}
	selectCmd = append(selectCmd, aid...)
	_, err = a.transmit(selectCmd)
	if err != nil {
		return 0, err
	}
	return int(a.channel), nil
}

func (a *echoAPDU) CloseLogicalChannel(channel []byte) error {
	_, err := a.transmit([]byte{0x00, 0x70, 0x80, a.channel, 0x00})
	return err
}

func main() {
	euicc, err := NewLibeuicc(NewEchoApdu())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer euicc.Free()

	fmt.Println(euicc.ProcessNotification(65, false))

	context, cancel := context.WithDeadline(context.Background(), time.Now().Add(10*time.Minute))
	err = euicc.DownloadProfile(context, &ActivationCode{
		SMDP:       "millicomelsalvador.validereachdpplus.com",
		MatchingId: "GENERICJOWMI-FAHTCU0-SKFMYPW6UIEFGRWC8GE933ITFAUVN63WMUVHFOWTS81",
	}, func(progress DownloadProgress, profileMetadata *ProfileMetadata, confirmDownloadChan chan bool, confirmationCodeChan chan string) {
		fmt.Println(progress, profileMetadata)
		if progress == DownloadProgressConfirmationCodeRequired {
			confirmationCodeChan <- "123456"
		}
		if progress == DownloadProgressConfirmDownload {
			confirmDownloadChan <- true
		}
	})
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("Download profile success")
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	cancel()
}
