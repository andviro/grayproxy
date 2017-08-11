package gelf

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"
)

// Chunk represent GELF UDP chunk
type Chunk []byte

var (
	gelfMagic  = []byte{0x1e, 0x0f}
	gzipMagic  = []byte{0x1f, 0x8b}
	zlibMagic0 = []byte{0x78, 0x01}
	zlibMagic1 = []byte{0x78, 0x9c}
	zlibMagic2 = []byte{0x78, 0xda}
)

// IsGELF returns true if the byte chunk starts with the GELF magic sequence
func (c Chunk) IsGELF() bool {
	return len(c) > 11 && bytes.Compare(c[:2], gelfMagic) == 0
}

// ID returns chunk ID
func (c Chunk) ID() string {
	return string(c[2:10])
}

// Sequence returns chunk sequence number and count
func (c Chunk) Sequence() (number, count int) {
	return int(c[10]), int(c[11])
}

// Body returns chunk payload
func (c Chunk) Body() []byte {
	return c[12:]
}

// Reader returns appropriate decoding reader for the chunk
func (c Chunk) Reader() (res io.Reader, err error) {
	m := c[:2]
	res = bytes.NewReader(c)
	switch {
	case bytes.Compare(m, zlibMagic0) == 0 || bytes.Compare(m, zlibMagic1) == 0 || bytes.Compare(m, zlibMagic2) == 0:
		return zlib.NewReader(res)
	case bytes.Compare(c[:2], gzipMagic) == 0:
		return gzip.NewReader(res)
	default:
		return
	}

}
