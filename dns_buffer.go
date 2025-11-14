package main

import (
	"encoding/binary"
	"errors"
	"strings"
)

// BytePacketBuffer helper for reading/writing DNS packet bytes (512-byte buffer)
type BytePacketBuffer struct {
	buf []byte
	pos int
}

func NewPacketBufferWithSize(size int) *BytePacketBuffer {
	return &BytePacketBuffer{
		buf: make([]byte, size),
		pos: 0,
	}
}

func NewPacketBuffer() *BytePacketBuffer {
	return NewPacketBufferWithSize(512)
}

func (b *BytePacketBuffer) Reset() {
	for i := range b.buf {
		b.buf[i] = 0
	}
	b.pos = 0
}

func (b *BytePacketBuffer) Read() (byte, error) {
	if b.pos >= len(b.buf) {
		return 0, errors.New("end of buffer")
	}
	v := b.buf[b.pos]
	b.pos++
	return v, nil
}

func (b *BytePacketBuffer) ReadUint16() (uint16, error) {
	if b.pos+2 > len(b.buf) {
		return 0, errors.New("short buffer for uint16")
	}
	v := binary.BigEndian.Uint16(b.buf[b.pos : b.pos+2])
	b.pos += 2
	return v, nil
}

func (b *BytePacketBuffer) ReadUint32() (uint32, error) {
	if b.pos+4 > len(b.buf) {
		return 0, errors.New("short buffer for uint32")
	}
	v := binary.BigEndian.Uint32(b.buf[b.pos : b.pos+4])
	b.pos += 4
	return v, nil
}

func (b *BytePacketBuffer) Write(bytes []byte) error {
	if b.pos+len(bytes) > len(b.buf) {
		return errors.New("buffer overflow write")
	}
	copy(b.buf[b.pos:], bytes)
	b.pos += len(bytes)
	return nil
}

func (b *BytePacketBuffer) WriteUint16(v uint16) error {
	if b.pos+2 > len(b.buf) {
		return errors.New("buffer overflow write u16")
	}
	binary.BigEndian.PutUint16(b.buf[b.pos:b.pos+2], v)
	b.pos += 2
	return nil
}

func (b *BytePacketBuffer) WriteUint32(v uint32) error {
	if b.pos+4 > len(b.buf) {
		return errors.New("buffer overflow write u32")
	}
	binary.BigEndian.PutUint32(b.buf[b.pos:b.pos+4], v)
	b.pos += 4
	return nil
}

func (b *BytePacketBuffer) Current() int { return b.pos }
func (b *BytePacketBuffer) Seek(pos int) error {
	if pos < 0 || pos >= len(b.buf) {
		return errors.New("invalid seek")
	}
	b.pos = pos
	return nil
}
func (b *BytePacketBuffer) Bytes() []byte { return b.buf[:b.pos] }
func (b *BytePacketBuffer) Len() int     { return b.pos }

// ReadQName reads a domain name, following pointers (RFC 1035) and returns full name
func (b *BytePacketBuffer) ReadQName() (string, error) {
	labels := []string{}
	visited := 0
	for {
		if visited > 256 {
			return "", errors.New("too many label jumps")
		}
		lenb, err := b.Read()
		if err != nil {
			return "", err
		}
		// pointer?
		if lenb&0xC0 == 0xC0 {
			offsetByte, err := b.Read()
			if err != nil {
				return "", err
			}
			offset := int(((uint16(lenb) ^ 0xC0) << 8) | uint16(offsetByte))
			cur := b.pos
			if err := b.Seek(offset); err != nil {
				return "", err
			}
			part, err := b.ReadQName()
			if err != nil {
				return "", err
			}
			labels = append(labels, part)
			if err := b.Seek(cur); err != nil {
				return "", err
			}
			break
		}
		if lenb == 0 {
			break
		}
		if b.pos+int(lenb) > len(b.buf) {
			return "", errors.New("label length overflow")
		}
		label := string(b.buf[b.pos : b.pos+int(lenb)])
		labels = append(labels, label)
		b.pos += int(lenb)
		visited++
	}
	return strings.Join(labels, "."), nil
}

// WriteQName writes a domain name without pointer compression
func (b *BytePacketBuffer) WriteQName(name string) error {
	if name == "" {
		return b.Write([]byte{0})
	}
	parts := strings.Split(name, ".")
	for _, p := range parts {
		if len(p) > 63 {
			return errors.New("label too long")
		}
		if err := b.Write([]byte{byte(len(p))}); err != nil {
			return err
		}
		if err := b.Write([]byte(p)); err != nil {
			return err
		}
	}
	return b.Write([]byte{0})
}
