package main

import (
	"errors"
	"net"
	"time"
)

type DnsRecord struct {
	Name  string
	Type  QType
	Class QClass
	TTL   uint32
	Data  []byte // raw RDATA
	// parsed helpers
	AData net.IP
	CName string
}

func (r *DnsRecord) Write(buf *BytePacketBuffer) error {
	if err := buf.WriteQName(r.Name); err != nil {
		return err
	}
	if err := buf.WriteUint16(uint16(r.Type)); err != nil {
		return err
	}
	if err := buf.WriteUint16(uint16(r.Class)); err != nil {
		return err
	}
	if err := buf.WriteUint32(r.TTL); err != nil {
		return err
	}
	if err := buf.WriteUint16(uint16(len(r.Data))); err != nil {
		return err
	}
	if len(r.Data) > 0 {
		if err := buf.Write(r.Data); err != nil {
			return err
		}
	}
	return nil
}

func ReadRecord(buf *BytePacketBuffer) (*DnsRecord, error) {
	name, err := buf.ReadQName()
	if err != nil { return nil, err }
	t, err := buf.ReadUint16()
	if err != nil { return nil, err }
	c, err := buf.ReadUint16()
	if err != nil { return nil, err }
	ttl, err := buf.ReadUint32()
	if err != nil { return nil, err }
	rdlen, err := buf.ReadUint16()
	if err != nil { return nil, err }

	if rdlen == 0 {
		return &DnsRecord{
			Name: name,
			Type: QType(t),
			Class: QClass(c),
			TTL: ttl,
			Data: []byte{},
		}, nil
	}

	if buf.pos+int(rdlen) > len(buf.buf) {
		return nil, errors.New("rdata too long")
	}
	rdata := make([]byte, rdlen)
	copy(rdata, buf.buf[buf.pos:buf.pos+int(rdlen)])
	buf.pos += int(rdlen)

	rec := &DnsRecord{
		Name: name,
		Type: QType(t),
		Class: QClass(c),
		TTL: ttl,
		Data: rdata,
	}
	switch rec.Type {
	case QTypeA:
		if len(rdata) == 4 {
			rec.AData = net.IPv4(rdata[0], rdata[1], rdata[2], rdata[3])
		}
	case QTypeAAAA:
		if len(rdata) == 16 {
			rec.AData = net.IP(rdata)
		}
	}
	return rec, nil
}

func NewARecord(name string, ipStr string, ttlSeconds uint32) (*DnsRecord, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, errors.New("invalid ip")
	}
	ip = ip.To4()
	if ip == nil {
		return nil, errors.New("not ipv4")
	}
	data := []byte{ip[0], ip[1], ip[2], ip[3]}
	rec := &DnsRecord{
		Name: name,
		Type: QTypeA,
		Class: QClassIN,
		TTL: ttlSeconds,
		Data: data,
		AData: ip,
	}
	return rec, nil
}

func (r *DnsRecord) ExpiryTime() time.Time {
	return time.Now().Add(time.Duration(r.TTL) * time.Second)
}
