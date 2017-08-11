package gelf

import (
	"bytes"
	"time"
)

type err string

func (e err) Error() string {
	return string(e)
}

// ErrTimeout is returned by Assembler if the current chunk came too late.
const ErrTimeout = err("Assembly time exceeded")

// ErrInvalidCount is returned by Assembler if the chunk sequence number and count do not match with the initial values
const ErrInvalidCount = err("Chunk out of sequence")

// ErrMessageTooLong is returned by Assembler if the message size limit
// exceeded.
const ErrMessageTooLong = err("Message too long")

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

// Update feeds the byte chunk to Assembler, returns ok when the message is
// complete.
func (a *Assembler) Update(chunk Chunk) (ok bool, err error) {
	if time.Now().After(a.deadline) {
		err = ErrTimeout
		return
	}
	num, count := chunk.Sequence()
	if a.fullMsg == nil {
		a.fullMsg = make([][]byte, count)
	}
	if count != len(a.fullMsg) || num >= count {
		err = ErrInvalidCount
		return
	}
	body := chunk.Body()

	if a.fullMsg[num] == nil {
		a.totalBytes += len(body)
		if a.totalBytes > a.maxMessageSize {
			err = ErrMessageTooLong
			return
		}
		a.fullMsg[num] = body
		a.processed++
	}
	ok = a.processed == len(a.fullMsg)
	return
}
