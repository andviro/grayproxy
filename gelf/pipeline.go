package gelf

import (
	"io"
	"time"
)

const periodicCleanup = 5 * time.Second

// Assemble consumes byte chunks from the input channel, usually passed from
// UDP server. It feeds de-chunked messages to the result channel.
func Assemble(chunks <-chan Chunk, maxMessageSize int, assembleTimeout time.Duration) <-chan Chunk {
	encodedMsgs := make(chan Chunk)
	go func() {
		defer close(encodedMsgs)
		assemblers := make(map[string]*Assembler)
		for {
			select {
			case chunk, ok := <-chunks:
				if !ok {
					return
				}
				if chunk.IsGELF() {
					cid := chunk.ID()
					a, ok := assemblers[cid]
					if !ok {
						a = NewAssembler(maxMessageSize, assembleTimeout)
						assemblers[cid] = a
					}
					if !a.Update(chunk) {
						continue
					}
					chunk = a.Bytes()
					delete(assemblers, cid)
				}
				encodedMsgs <- chunk
			case <-time.After(periodicCleanup):
				for k, v := range assemblers {
					if v.Expired() {
						delete(assemblers, k)
					}
				}
			}
		}
	}()
	return encodedMsgs
}

// Extract applies decompression to byte messages if nessessary.
func Extract(encodedMsgs <-chan Chunk, decompressSizeLimit int) <-chan []byte {
	messages := make(chan []byte)
	go func() {
		defer close(messages)
		for msg := range encodedMsgs {
			r, err := msg.Reader()
			if err != nil {
				continue
			}
			buf := make([]byte, decompressSizeLimit)
			n, err := r.Read(buf)
			if err != nil && err != io.EOF {
				continue
			}
			messages <- buf[:n]
		}
	}()
	return messages
}
