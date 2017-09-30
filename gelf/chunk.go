package gelf

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"
	"io/ioutil"
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
	return len(c) > 11 && bytes.Equal(c[:2], gelfMagic)
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

// Data returns appropriately decoded message bytes
func (c Chunk) Data(decompressSizeLimit int) (res []byte, err error) {
	var r io.Reader
	m := c[:2]
	switch {
	case bytes.Equal(m, zlibMagic0) || bytes.Equal(m, zlibMagic1) || bytes.Equal(m, zlibMagic2):
		if r, err = zlib.NewReader(bytes.NewReader(c)); err != nil {
			return
		}
	case bytes.Equal(m, gzipMagic):
		if r, err = gzip.NewReader(bytes.NewReader(c)); err != nil {
			return
		}
	default:
		return []byte(c), nil
	}
	if decompressSizeLimit > 0 {
		r = io.LimitReader(r, int64(decompressSizeLimit))
	}
	return ioutil.ReadAll(r)
}
