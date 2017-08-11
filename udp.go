package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/andviro/grayproxy/gelf"
)

type udpListener struct {
	Address                                                            string
	MaxChunkSize, MaxMessageSize, DecompressSizeLimit, AssembleTimeout int
}

func (in *udpListener) String() string {
	return fmt.Sprint(*in)
}

func (in *udpListener) Set(val string) error {
	n, _ := fmt.Sscan(strings.TrimPrefix(val, "udp://"), &in.Address, &in.MaxChunkSize, &in.MaxMessageSize, &in.DecompressSizeLimit, &in.AssembleTimeout)
	if n == 0 {
		return errors.New("empty input description")
	}
	if in.MaxChunkSize == 0 {
		in.MaxChunkSize = 8192
	}
	if in.MaxMessageSize == 0 {
		in.MaxChunkSize = 128 * 1024
	}
	if in.DecompressSizeLimit == 0 {
		in.DecompressSizeLimit = 1024 * 1024
	}
	if in.AssembleTimeout == 0 {
		in.AssembleTimeout = 1000
	}
	return nil
}

func (in *udpListener) listen(dest chan gelf.Chunk) (err error) {
	chunks := make(chan gelf.Chunk)
	defer close(chunks)

	decodedMsgs := gelf.Extract(gelf.Assemble(chunks, in.MaxMessageSize, time.Millisecond*time.Duration(in.AssembleTimeout)), in.DecompressSizeLimit)
	go func() {
		for msg := range decodedMsgs {
			dest <- msg
		}
	}()

	l, err := net.ListenPacket("udp", in.Address)
	if err != nil {
		return errors.Wrap(err, "listening on UDP port")
	}
	buf := make([]byte, in.MaxChunkSize)
	for {
		n, _, err := l.ReadFrom(buf)
		if err != nil {
			return errors.Wrap(err, "reading UDP packet")
		}
		chunks <- buf[:n]
	}
}
