package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	DefaultPort   = 2053
	UpstreamDNS   = "8.8.8.8:53"
	MaxPacketSize = 512
)

// DnsServer represents the DNS server
type DnsServer struct {
	port     int
	cache    *DnsCache
	resolver *DnsResolver
	udpConn  *net.UDPConn
}

// NewDnsServer creates a new DNS server
func NewDnsServer(port int) *DnsServer {
	return &DnsServer{
		port:     port,
		cache:    NewDnsCache(),
		resolver: NewDnsResolver(UpstreamDNS),
	}
}

// Start starts UDP and TCP DNS servers
func (s *DnsServer) Start() error {
	// -----------------------
	// UDP SETUP (port 2053)
	// -----------------------
	udpAddr := &net.UDPAddr{
		Port: s.port,
		IP:   net.ParseIP("0.0.0.0"),
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to bind UDP port %d: %w", s.port, err)
	}
	s.udpConn = udpConn

	log.Printf("🚀 DNS UDP Server started on 0.0.0.0:%d", s.port)

	// -----------------------
	// TCP SETUP (port 2053)
	// -----------------------
	go s.startTCPServer()

	log.Printf("📡 Upstream DNS: %s", UpstreamDNS)
	log.Printf("💾 Cache initialized")
	log.Println("Ready to handle queries...")

	// Handle UDP requests
	return s.handleUDPRequests()
}

// ---------------------------
// UDP REQUEST HANDLER
// ---------------------------
func (s *DnsServer) handleUDPRequests() error {
	buffer := make([]byte, MaxPacketSize)

	for {
		n, clientAddr, err := s.udpConn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("❌ Error reading UDP: %v", err)
			continue
		}
		go s.processDNSQuery(buffer[:n], clientAddr)
	}
}

// ---------------------------
// TCP SERVER (NEW)
// ---------------------------
func (s *DnsServer) startTCPServer() {
	addr := fmt.Sprintf("0.0.0.0:%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("❌ TCP listen failed: %v", err)
		return
	}

	log.Printf("🚀 DNS TCP Server started on 0.0.0.0:%d", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("❌ TCP accept error: %v", err)
			continue
		}
		go s.handleTCPConnection(conn)
	}
}

// ---------------------------
// TCP CONNECTION HANDLER (NEW)
// ---------------------------
func (s *DnsServer) handleTCPConnection(conn net.Conn) {
	defer conn.Close()

	// First 2 bytes = request length
	lenBuf := make([]byte, 2)
	_, err := conn.Read(lenBuf)
	if err != nil {
		log.Printf("❌ TCP failed reading length: %v", err)
		return
	}

	msgLen := int(lenBuf[0])<<8 | int(lenBuf[1])
	if msgLen <= 0 {
		log.Printf("❌ TCP invalid length")
		return
	}

	msg := make([]byte, msgLen)
	_, err = conn.Read(msg)
	if err != nil {
		log.Printf("❌ TCP failed reading message: %v", err)
		return
	}

	// Parse request
	packet, err := FromBytes(msg)
	if err != nil {
		log.Printf("❌ TCP parse failed: %v", err)
		return
	}

	clientAddr := conn.RemoteAddr().String()
	startTime := time.Now()

	// Process request (same code used for UDP)
	responsePacket := s.buildResponse(packet, clientAddr)

	// Encode response
	responseBytes, err := responsePacket.ToBytes()
	if err != nil {
		log.Printf("❌ TCP encode failed: %v", err)
		return
	}

	// TCP requires sending length prefix
	respLen := []byte{byte(len(responseBytes) >> 8), byte(len(responseBytes))}
	conn.Write(respLen)
	conn.Write(responseBytes)

	log.Printf("📤 TCP Response sent to %s in %v (size: %d bytes)",
		clientAddr, time.Since(startTime), len(responseBytes))
}

// ---------------------------
// SHARED LOGIC FOR UDP & TCP
// ---------------------------
func (s *DnsServer) processDNSQuery(data []byte, clientAddr *net.UDPAddr) {
	packet, err := FromBytes(data)
	if err != nil {
		log.Printf("❌ Failed to parse UDP request: %v", err)
		return
	}

	responsePacket := s.buildResponse(packet, clientAddr.String())

	responseBytes, err := responsePacket.ToBytes()
	if err != nil {
		log.Printf("❌ Failed to encode response: %v", err)
		return
	}

	s.udpConn.WriteToUDP(responseBytes, clientAddr)
}

func (s *DnsServer) buildResponse(requestPacket *DnsPacket, client string) *DnsPacket {
	startTime := time.Now()
	responsePacket := NewDnsPacket()

	// Copy header
	responsePacket.Header.ID = requestPacket.Header.ID
	responsePacket.Header.Response = true
	responsePacket.Header.Opcode = requestPacket.Header.Opcode
	responsePacket.Header.RecursionDesired = requestPacket.Header.RecursionDesired
	responsePacket.Header.RecursionAvailable = true

	// Copy questions
	responsePacket.Questions = requestPacket.Questions

	// Process questions
	for _, q := range requestPacket.Questions {
		log.Printf("📥 Query from %s: %s [%s]", client, q.Name, q.QType.String())

		// Cache hit?
		if cached, ok := s.cache.Get(q.Name, q.QType); ok {
			log.Printf("✅ Cache HIT: %s [%s]", q.Name, q.QType.String())
			responsePacket.Answers = append(responsePacket.Answers, cached)
			continue
		}

		// Cache miss → upstream
		log.Printf("❌ Cache MISS: %s [%s] - querying upstream", q.Name, q.QType.String())

		upstreamPacket, err := s.resolver.RecursiveLookup(q.Name, q.QType)
		if err != nil {
			log.Printf("❌ Upstream error: %v", err)
			responsePacket.Header.RESCODE = SERVFAIL
			continue
		}

		responsePacket.Answers = append(responsePacket.Answers, upstreamPacket.Answers...)
		responsePacket.Authorities = append(responsePacket.Authorities, upstreamPacket.Authorities...)
		responsePacket.Resources = append(responsePacket.Resources, upstreamPacket.Resources...)

		s.cache.PutMultiple(upstreamPacket.Answers)

		log.Printf("✅ Upstream resolved: %d answers", len(upstreamPacket.Answers))
	}

	log.Printf("⏱️ Processed query in %v", time.Since(startTime))
	return responsePacket
}

func (s *DnsServer) Stop() {
	if s.udpConn != nil {
		log.Println("🛑 Shutting down DNS server...")
		s.udpConn.Close()
	}
}

func (s *DnsServer) PrintStats() {
	log.Printf("📊 Cache Stats: %s", s.cache.Stats())
}

func main() {
	server := NewDnsServer(DefaultPort)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()

	// Print cache stats every 30 seconds
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			server.PrintStats()
		}
	}()

	<-sigChan
	server.Stop()
}
