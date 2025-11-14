package main

import "fmt"

type QType uint16
type QClass uint16

const (
	QTypeA     QType = 1
	QTypeNS    QType = 2
	QTypeCNAME QType = 5
	QTypeSOA   QType = 6
	QTypeMX    QType = 15
	QTypeAAAA  QType = 28
)

const (
	QClassIN QClass = 1
)

func (qt QType) String() string {
	switch qt {
	case QTypeA:
		return "A"
	case QTypeCNAME:
		return "CNAME"
	case QTypeAAAA:
		return "AAAA"
	case QTypeMX:
		return "MX"
	default:
		return fmt.Sprintf("TYPE%d", uint16(qt))
	}
}

type DnsQuestion struct {
	Name   string
	QType  QType
	QClass QClass
}

func (q *DnsQuestion) Read(buf *BytePacketBuffer) error {
	name, err := buf.ReadQName()
	if err != nil {
		return err
	}
	q.Name = name
	t, err := buf.ReadUint16()
	if err != nil { return err }
	c, err := buf.ReadUint16()
	if err != nil { return err }
	q.QType = QType(t)
	q.QClass = QClass(c)
	return nil
}

func (q *DnsQuestion) Write(buf *BytePacketBuffer) error {
	if err := buf.WriteQName(q.Name); err != nil {
		return err
	}
	if err := buf.WriteUint16(uint16(q.QType)); err != nil {
		return err
	}
	if err := buf.WriteUint16(uint16(q.QClass)); err != nil {
		return err
	}
	return nil
}
