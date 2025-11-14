package main

import (
    "io"
    "log"
    "net"
	"fmt"
)

func (s *DnsServer) StartTCP() error {
    addr := fmt.Sprintf("0.0.0.0:%d", s.port)

    ln, err := net.Listen("tcp", addr)
    if err != nil {
        return fmt.Errorf("❌ TCP listen failed: %w", err)
    }

    log.Printf("🔌 TCP DNS Server started on %s", addr)

    go func() {
        for {
            conn, err := ln.Accept()
            if err != nil {
                log.Printf("❌ TCP accept failed: %v", err)
                continue
            }
            go s.handleTCPConn(conn)
        }
    }()

    return nil
}

func (s *DnsServer) handleTCPConn(conn net.Conn) {
    defer conn.Close()

    // TCP DNS packets have a 2-byte length prefix
    lengthBuf := make([]byte, 2)
    if _, err := io.ReadFull(conn, lengthBuf); err != nil {
        return
    }
    size := int(lengthBuf[0])<<8 | int(lengthBuf[1])

    // Read full DNS message
    buf := make([]byte, size)
    if _, err := io.ReadFull(conn, buf); err != nil {
        return
    }

    // Parse DNS request
    requestPacket, err := FromBytes(buf)
    if err != nil {
        log.Printf("❌ TCP parse error: %v", err)
        return
    }

    // Prepare response using the SAME logic used for UDP
    responsePacket := s.makeResponseFromPacket(requestPacket)

    respBytes, err := responsePacket.ToBytes()
    if err != nil {
        log.Printf("❌ TCP encode error: %v", err)
        return
    }

    // Write length prefix
    header := []byte{byte(len(respBytes) >> 8), byte(len(respBytes))}
    conn.Write(header)
    conn.Write(respBytes)

    log.Printf("📤 TCP response sent (size: %d bytes)", len(respBytes))
}

// Shared response creation logic (factored out)
func (s *DnsServer) makeResponseFromPacket(requestPacket *DnsPacket) *DnsPacket {
    responsePacket := NewDnsPacket()
    responsePacket.Header.ID = requestPacket.Header.ID
    responsePacket.Header.Response = true
    responsePacket.Header.Opcode = requestPacket.Header.Opcode
    responsePacket.Header.RecursionDesired = requestPacket.Header.RecursionDesired
    responsePacket.Header.RecursionAvailable = true

    responsePacket.Questions = requestPacket.Questions

    for _, question := range requestPacket.Questions {
        if cached, found := s.cache.Get(question.Name, question.QType); found {
            responsePacket.Answers = append(responsePacket.Answers, cached)
            responsePacket.Header.RESCODE = NOERROR
            continue
        }

        upstreamPacket, err := s.resolver.RecursiveLookup(question.Name, question.QType)
        if err != nil {
            responsePacket.Header.RESCODE = SERVFAIL
            continue
        }

        responsePacket.Answers = append(responsePacket.Answers, upstreamPacket.Answers...)
        responsePacket.Authorities = append(responsePacket.Authorities, upstreamPacket.Authorities...)
        responsePacket.Resources = append(responsePacket.Resources, upstreamPacket.Resources...)

        s.cache.PutMultiple(upstreamPacket.Answers)
        responsePacket.Header.RESCODE = upstreamPacket.Header.RESCODE
    }

    return responsePacket
}
