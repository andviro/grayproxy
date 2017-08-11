package gelf

import (
	"bytes"
	"time"
)

// Assembler provides GELF message de-chunking
type Assembler struct {
	deadline                   time.Time
	fullMsg                    [][]byte
	processed                  int
	totalBytes, maxMessageSize int
}

// NewAssembler returns empty Assembler with maximum message size and duration
// specified.
func NewAssembler(maxMessageSize int, timeout time.Duration) (res *Assembler) {
	res = new(Assembler)
	res.deadline = time.Now().Add(timeout)
	res.maxMessageSize = maxMessageSize
	return
}

// Bytes returns message bytes, not nessesarily fully assembled
func (a *Assembler) Bytes() []byte {
	return bytes.Join(a.fullMsg, nil)
}

// Expired returns true if first chunk is too old
func (a *Assembler) Expired() bool {
	return time.Now().After(a.deadline)
}

// Update feeds the byte chunk to Assembler, returns ok when the message is
// complete.
func (a *Assembler) Update(chunk Chunk) bool {
	num, count := chunk.Sequence()
	if a.fullMsg == nil {
		a.fullMsg = make([][]byte, count)
	}
	if count != len(a.fullMsg) || num >= count {
		return false
	}
	body := chunk.Body()

	if a.fullMsg[num] == nil {
		a.totalBytes += len(body)
		if a.totalBytes > a.maxMessageSize {
			return false
		}
		a.fullMsg[num] = body
		a.processed++
	}
	return a.processed == len(a.fullMsg)
}
