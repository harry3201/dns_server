package main

import (
	"net"
	"time"
	"log"
)

// DnsResolver handles forwarding queries to an upstream server
type DnsResolver struct {
	upstream string
	timeout  time.Duration
}

func NewDnsResolver(upstream string) *DnsResolver {
	return &DnsResolver{
		upstream: upstream,
		timeout:  3 * time.Second,
	}
}

// RecursiveLookup forwards the raw query and returns the parsed upstream packet
func (r *DnsResolver) RecursiveLookup(name string, qtype QType) (*DnsPacket, error) {
	// Build question-only packet (we can forward original query bytes instead).
	// Simpler: create a UDP connection to upstream and forward the raw packet
	// but we don't have the original raw bytes here; instead create a minimal query.

	// We'll create a very small query: but easiest approach — use net.Dial and send a properly formatted query:
	// For simplicity we will forward a new query built from scratch.
	// Build a DnsPacket
	pkt := NewDnsPacket()
	// random ID
	pkt.Header.ID = uint16(time.Now().UnixNano() & 0xffff)
	pkt.Header.RecursionDesired = true
	pkt.Questions = append(pkt.Questions, &DnsQuestion{
		Name: name,
		QType: qtype,
		QClass: QClassIN,
	})

	// serialize
	raw, err := pkt.ToBytes()
	if err != nil {
		return nil, err
	}

	// send to upstream
	raddr, err := net.ResolveUDPAddr("udp", r.upstream)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(r.timeout))

	if _, err := conn.Write(raw); err != nil {
		return nil, err
	}

	resp := make([]byte, 2048)
	n, err := conn.Read(resp)
	if err != nil {
		return nil, err
	}
	respCopy := make([]byte, n)
	copy(respCopy, resp[:n])

	// parse upstream response
	upPkt, err := FromBytes(respCopy)
	if err != nil {
		// log parse error but still return raw payload as nil error
		log.Printf("warning: failed to parse upstream response: %v", err)
		return nil, err
	}
	return upPkt, nil
}
