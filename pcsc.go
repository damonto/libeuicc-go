package libeuicc

import (
	"errors"

	"github.com/ElMostafaIdrassi/goscard"
)

type PCSC interface {
	APDU
	ListReaders() ([]string, error)
	SetReader(reader string)
}

type PCSCReader struct {
	context goscard.Context
	card    goscard.Card
	channel byte
	reader  string
}

func NewPCSCReader() (PCSC, error) {
	pcsc := &PCSCReader{}
	if err := goscard.Initialize(goscard.NewDefaultLogger(goscard.LogLevelNone)); err != nil {
		return nil, err
	}

	context, _, err := goscard.NewContext(goscard.SCardScopeSystem, nil, nil)
	if err != nil {
		return nil, err
	}
	pcsc.context = context
	readers, err := pcsc.ListReaders()
	if err != nil {
		return nil, err
	}
	pcsc.SetReader(readers[0])
	return pcsc, nil
}

func (p *PCSCReader) ListReaders() ([]string, error) {
	readers, _, err := p.context.ListReaders(nil)
	if err != nil {
		return nil, err
	}
	if len(readers) == 0 {
		return nil, errors.New("no readers found")
	}
	return readers, nil
}

func (p *PCSCReader) SetReader(reader string) {
	p.reader = reader
}

func (p *PCSCReader) Connect() error {
	card, _, err := p.context.Connect(p.reader, goscard.SCardShareExclusive, goscard.SCardProtocolT0)
	if err != nil {
		return err
	}
	p.card = card
	_, err = p.Transmit([]byte{0x80, 0xAA, 0x00, 0x00, 0x0A, 0xA9, 0x08, 0x81, 0x00, 0x82, 0x01, 0x01, 0x83, 0x01, 0x07})
	return err
}

func (p *PCSCReader) Disconnect() error {
	defer goscard.Finalize()
	if _, err := p.card.Disconnect(goscard.SCardLeaveCard); err != nil {
		logger.Error("apdu error disconnecting card", err)
		return err
	}
	if _, err := p.context.Release(); err != nil {
		logger.Error("apdu error releasing context", err)
		return err
	}
	return nil
}

func (p *PCSCReader) Transmit(command []byte) ([]byte, error) {
	resp, _, err := p.card.Transmit(&goscard.SCardIoRequestT0, command, nil)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *PCSCReader) OpenLogicalChannel(aid []byte) (int, error) {
	resp, err := p.Transmit([]byte{0x00, 0x70, 0x00, 0x00, 0x01})
	if err != nil {
		return 0, err
	}
	if resp[1] != 0x90 {
		return 0, errors.New("failed to open logical channel")
	}

	p.channel = resp[0]
	command := []byte{p.channel, 0xA4, 0x04, 0x00, byte(len(aid))}
	command = append(command, aid...)
	resp, err = p.Transmit(command)
	if err != nil {
		return 0, err
	}
	if resp[0] != 0x90 && resp[0] != 0x61 {
		return 0, errors.New("failed to select AID")
	}
	return int(p.channel), nil
}

func (p *PCSCReader) CloseLogicalChannel(channel []byte) error {
	command := []byte{0x00, 0x70, 0x80, channel[0], 0x00}
	_, err := p.Transmit(command)
	return err
}
