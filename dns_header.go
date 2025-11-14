package main

type RCode uint8

const (
	NOERROR RCode = 0
	FORMERR RCode = 1
	SERVFAIL RCode = 2
	NXDOMAIN RCode = 3
	NOTIMPL  RCode = 4
	REFUSED  RCode = 5
)

type DnsHeader struct {
	ID               uint16
	Response         bool
	Opcode           uint8
	Authoritative    bool
	Truncated        bool
	RecursionDesired bool
	RecursionAvailable bool
	Z                uint8
	RESCODE          RCode

	QDCount uint16
	ANCount uint16
	NSCount uint16
	ARCount uint16
}

func NewDnsHeader() *DnsHeader {
	return &DnsHeader{
		Opcode: 0,
		QDCount: 0,
		ANCount: 0,
		NSCount: 0,
		ARCount: 0,
		RESCODE: NOERROR,
	}
}

func (h *DnsHeader) Read(buf *BytePacketBuffer) error {
	id, err := buf.ReadUint16()
	if err != nil { return err }
	h.ID = id
	flags, err := buf.ReadUint16()
	if err != nil { return err }
	h.Response = (flags & 0x8000) != 0
	h.Opcode = uint8((flags >> 11) & 0x0F)
	h.Authoritative = (flags & 0x0400) != 0
	h.Truncated = (flags & 0x0200) != 0
	h.RecursionDesired = (flags & 0x0100) != 0
	h.RecursionAvailable = (flags & 0x0080) != 0
	h.Z = uint8((flags >> 4) & 0x7)
	h.RESCODE = RCode(flags & 0x000F)

	qd, _ := buf.ReadUint16()
	an, _ := buf.ReadUint16()
	ns, _ := buf.ReadUint16()
	ar, _ := buf.ReadUint16()
	h.QDCount = qd
	h.ANCount = an
	h.NSCount = ns
	h.ARCount = ar
	return nil
}

func (h *DnsHeader) Write(buf *BytePacketBuffer) error {
	if err := buf.WriteUint16(h.ID); err != nil { return err }
	var flags uint16 = 0
	if h.Response { flags |= 0x8000 }
	flags |= (uint16(h.Opcode&0x0F) << 11)
	if h.Authoritative { flags |= 0x0400 }
	if h.Truncated { flags |= 0x0200 }
	if h.RecursionDesired { flags |= 0x0100 }
	if h.RecursionAvailable { flags |= 0x0080 }
	flags |= (uint16(h.Z&0x7) << 4)
	flags |= uint16(h.RESCODE & 0x0F)

	if err := buf.WriteUint16(flags); err != nil { return err }
	if err := buf.WriteUint16(h.QDCount); err != nil { return err }
	if err := buf.WriteUint16(h.ANCount); err != nil { return err }
	if err := buf.WriteUint16(h.NSCount); err != nil { return err }
	if err := buf.WriteUint16(h.ARCount); err != nil { return err }
	return nil
}
