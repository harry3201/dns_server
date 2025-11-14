package main

type DnsPacket struct {
	Header     *DnsHeader
	Questions  []*DnsQuestion
	Answers    []*DnsRecord
	Authorities []*DnsRecord
	Resources  []*DnsRecord
}

func NewDnsPacket() *DnsPacket {
	return &DnsPacket{
		Header: NewDnsHeader(),
		Questions: make([]*DnsQuestion, 0),
		Answers: make([]*DnsRecord, 0),
		Authorities: make([]*DnsRecord, 0),
		Resources: make([]*DnsRecord, 0),
	}
}

func FromBytes(data []byte) (*DnsPacket, error) {
	buf := NewPacketBufferWithSize(len(data))
	copy(buf.buf, data)
	buf.pos = 0

	p := NewDnsPacket()
	if err := p.Header.Read(buf); err != nil {
		return nil, err
	}

	for i := 0; i < int(p.Header.QDCount); i++ {
		q := &DnsQuestion{}
		if err := q.Read(buf); err != nil {
			return nil, err
		}
		p.Questions = append(p.Questions, q)
	}

	for i := 0; i < int(p.Header.ANCount); i++ {
		r, err := ReadRecord(buf)
		if err != nil {
			return nil, err
		}
		p.Answers = append(p.Answers, r)
	}

	for i := 0; i < int(p.Header.NSCount); i++ {
		r, err := ReadRecord(buf)
		if err != nil {
			return nil, err
		}
		p.Authorities = append(p.Authorities, r)
	}

	for i := 0; i < int(p.Header.ARCount); i++ {
		r, err := ReadRecord(buf)
		if err != nil {
			return nil, err
		}
		p.Resources = append(p.Resources, r)
	}
	return p, nil
}

func (p *DnsPacket) ToBytes() ([]byte, error) {
	buf := NewPacketBuffer()
	p.Header.QDCount = uint16(len(p.Questions))
	p.Header.ANCount = uint16(len(p.Answers))
	p.Header.NSCount = uint16(len(p.Authorities))
	p.Header.ARCount = uint16(len(p.Resources))

	if err := p.Header.Write(buf); err != nil {
		return nil, err
	}
	for _, q := range p.Questions {
		if err := q.Write(buf); err != nil {
			return nil, err
		}
	}
	for _, a := range p.Answers {
		if err := a.Write(buf); err != nil {
			return nil, err
		}
	}
	for _, a := range p.Authorities {
		if err := a.Write(buf); err != nil {
			return nil, err
		}
	}
	for _, a := range p.Resources {
		if err := a.Write(buf); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}
