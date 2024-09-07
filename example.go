package main

import (
	"fmt"
	"log"
	"net"

	"encoding/hex"
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
	log.Println("Transmit", hex.EncodeToString(command))
	resp, err := a.transmit(command)
	if err != nil {
		return nil, err
	}
	log.Println("Received", hex.EncodeToString(resp))
	return resp, nil
}

func (a *echoAPDU) OpenLogicalChannel(aid []byte) (int, error) {
	logicalChannelResp, err := a.transmit([]byte{0x00, 0x70, 0x00, 0x00, 0x01})
	if err != nil {
		return 0, err
	}
	a.channel = logicalChannelResp[0]
	log.Println("OpenLogicalChannel", a.channel)
	selectCmd := []byte{a.channel, 0xA4, 0x04, 0x00, byte(len(aid))}
	selectCmd = append(selectCmd, aid...)
	resp, err := a.transmit(selectCmd)
	if err != nil {
		return 0, err
	}
	log.Println("Select", hex.EncodeToString(resp))
	return int(a.channel), nil
}

func (a *echoAPDU) CloseLogicalChannel(channel []byte) error {
	log.Println("CloseLogicalChannel", channel)
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

	fmt.Println(euicc.GetEid())
	fmt.Println(euicc.GetEuiccInfo2())
	ns, err := euicc.GetNotifications()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, n := range ns {
		fmt.Println(n.SeqNumber, n.ProfileManagementOperation, n.NotificationAddress, n.ICCID)
	}
}
