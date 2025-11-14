package main

import (
	"net"
	"testing"
	"time"
)

// TestBytePacketBuffer tests the buffer operations
func TestBytePacketBuffer(t *testing.T) {
	buf := NewBytePacketBuffer()

	// Test write/read u8
	if err := buf.writeU8(42); err != nil {
		t.Fatalf("writeU8 failed: %v", err)
	}
	buf.seek(0)
	val, err := buf.read()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}

	// Test write/read u16
	buf = NewBytePacketBuffer()
	if err := buf.writeU16(0x1234); err != nil {
		t.Fatalf("writeU16 failed: %v", err)
	}
	buf.seek(0)
	val16, err := buf.readU16()
	if err != nil {
		t.Fatalf("readU16 failed: %v", err)
	}
	if val16 != 0x1234 {
		t.Errorf("Expected 0x1234, got 0x%04x", val16)
	}

	// Test write/read u32
	buf = NewBytePacketBuffer()
	if err := buf.writeU32(0x12345678); err != nil {
		t.Fatalf("writeU32 failed: %v", err)
	}
	buf.seek(0)
	val32, err := buf.readU32()
	if err != nil {
		t.Fatalf("readU32 failed: %v", err)
	}
	if val32 != 0x12345678 {
		t.Errorf("Expected 0x12345678, got 0x%08x", val32)
	}
}

// TestQNameEncoding tests domain name encoding/decoding
func TestQNameEncoding(t *testing.T) {
	buf := NewBytePacketBuffer()

	// Write domain name
	domain := "google.com"
	if err := buf.writeQName(domain); err != nil {
		t.Fatalf("writeQName failed: %v", err)
	}

	// Read it back
	buf.seek(0)
	decoded, err := buf.readQName()
	if err != nil {
		t.Fatalf("readQName failed: %v", err)
	}

	if decoded != domain {
		t.Errorf("Expected %s, got %s", domain, decoded)
	}
}

// TestDnsHeader tests header encoding/decoding
func TestDnsHeader(t *testing.T) {
	header := NewDnsHeader()
	header.ID = 1234
	header.Response = true
	header.RecursionDesired = true
	header.RecursionAvailable = true
	header.Questions = 1
	header.Answers = 2

	// Write to buffer
	buf := NewBytePacketBuffer()
	if err := header.Write(buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read back
	buf.seek(0)
	decoded := NewDnsHeader()
	if err := decoded.Read(buf); err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Verify
	if decoded.ID != header.ID {
		t.Errorf("ID mismatch: expected %d, got %d", header.ID, decoded.ID)
	}
	if decoded.Response != header.Response {
		t.Errorf("Response flag mismatch")
	}
	if decoded.Questions != header.Questions {
		t.Errorf("Questions count mismatch")
	}
	if decoded.Answers != header.Answers {
		t.Errorf("Answers count mismatch")
	}
}

// TestDnsQuestion tests question encoding/decoding
func TestDnsQuestion(t *testing.T) {
	question := NewDnsQuestion("example.com", A)

	// Write to buffer
	buf := NewBytePacketBuffer()
	if err := question.Write(buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read back
	buf.seek(0)
	decoded := &DnsQuestion{}
	if err := decoded.Read(buf); err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Verify
	if decoded.Name != question.Name {
		t.Errorf("Name mismatch: expected %s, got %s", question.Name, decoded.Name)
	}
	if decoded.QType != question.QType {
		t.Errorf("QType mismatch: expected %v, got %v", question.QType, decoded.QType)
	}
}

// TestARecord tests A record encoding/decoding
func TestARecord(t *testing.T) {
	record := NewDnsRecord()
	record.Domain = "example.com"
	record.QType = A
	record.TTL = 300
	record.Addr = net.ParseIP("192.168.1.1")

	// Write to buffer
	buf := NewBytePacketBuffer()
	if err := record.Write(buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read back
	buf.seek(0)
	decoded := NewDnsRecord()
	if err := decoded.Read(buf); err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Verify
	if decoded.Domain != record.Domain {
		t.Errorf("Domain mismatch")
	}
	if decoded.QType != record.QType {
		t.Errorf("QType mismatch")
	}
	if decoded.TTL != record.TTL {
		t.Errorf("TTL mismatch")
	}
	if !decoded.Addr.Equal(record.Addr) {
		t.Errorf("Address mismatch: expected %s, got %s", record.Addr, decoded.Addr)
	}
}

// TestCNameRecord tests CNAME record encoding/decoding
func TestCNameRecord(t *testing.T) {
	record := NewDnsRecord()
	record.Domain = "www.example.com"
	record.QType = CNAME
	record.TTL = 300
	record.CName = "example.com"

	// Write to buffer
	buf := NewBytePacketBuffer()
	if err := record.Write(buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read back
	buf.seek(0)
	decoded := NewDnsRecord()
	if err := decoded.Read(buf); err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Verify
	if decoded.CName != record.CName {
		t.Errorf("CNAME mismatch: expected %s, got %s", record.CName, decoded.CName)
	}
}

// TestDnsPacket tests complete packet encoding/decoding
func TestDnsPacket(t *testing.T) {
	// Create packet
	packet := NewDnsPacket()
	packet.Header.ID = 1234
	packet.Header.RecursionDesired = true
	
	// Add question
	question := NewDnsQuestion("example.com", A)
	packet.Questions = append(packet.Questions, question)
	
	// Add answer
	answer := NewDnsRecord()
	answer.Domain = "example.com"
	answer.QType = A
	answer.TTL = 300
	answer.Addr = net.ParseIP("93.184.216.34")
	packet.Answers = append(packet.Answers, answer)

	// Encode to bytes
	bytes, err := packet.ToBytes()
	if err != nil {
		t.Fatalf("ToBytes failed: %v", err)
	}

	// Decode from bytes
	decoded, err := FromBytes(bytes)
	if err != nil {
		t.Fatalf("FromBytes failed: %v", err)
	}

	// Verify
	if decoded.Header.ID != packet.Header.ID {
		t.Errorf("ID mismatch")
	}
	if len(decoded.Questions) != len(packet.Questions) {
		t.Errorf("Questions count mismatch")
	}
	if len(decoded.Answers) != len(packet.Answers) {
		t.Errorf("Answers count mismatch")
	}
	if decoded.Questions[0].Name != question.Name {
		t.Errorf("Question name mismatch")
	}
	if !decoded.Answers[0].Addr.Equal(answer.Addr) {
		t.Errorf("Answer address mismatch")
	}
}

// TestCache tests DNS caching
func TestCache(t *testing.T) {
	cache := NewDnsCache()

	// Create a record
	record := NewDnsRecord()
	record.Domain = "example.com"
	record.QType = A
	record.TTL = 1 // 1 second TTL for testing
	record.Addr = net.ParseIP("93.184.216.34")

	// Put in cache
	cache.Put(record)

	// Get from cache
	cached, found := cache.Get("example.com", A)
	if !found {
		t.Error("Record not found in cache")
	}
	if !cached.Addr.Equal(record.Addr) {
		t.Error("Cached record mismatch")
	}

	// Wait for expiry
	time.Sleep(2 * time.Second)

	// Should be expired
	_, found = cache.Get("example.com", A)
	if found {
		t.Error("Record should have expired")
	}
}

// TestQueryTypeConversion tests query type conversion
func TestQueryTypeConversion(t *testing.T) {
	tests := []struct {
		num      uint16
		expected QueryType
	}{
		{1, A},
		{5, CNAME},
		{15, MX},
		{28, AAAA},
		{999, UNKNOWN},
	}

	for _, test := range tests {
		result := QueryTypeFromNum(test.num)
		if result != test.expected {
			t.Errorf("QueryTypeFromNum(%d) = %v, expected %v", test.num, result, test.expected)
		}
	}
}

// TestSplitDomain tests domain label splitting
func TestSplitDomain(t *testing.T) {
	tests := []struct {
		domain   string
		expected []string
	}{
		{"google.com", []string{"google", "com"}},
		{"www.example.com", []string{"www", "example", "com"}},
		{"", []string{}},
		{"single", []string{"single"}},
	}

	for _, test := range tests {
		result := splitDomain(test.domain)
		if len(result) != len(test.expected) {
			t.Errorf("splitDomain(%s) length mismatch", test.domain)
			continue
		}
		for i := range result {
			if result[i] != test.expected[i] {
				t.Errorf("splitDomain(%s)[%d] = %s, expected %s", 
					test.domain, i, result[i], test.expected[i])
			}
		}
	}
}
