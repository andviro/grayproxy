package udp

import (
	"net"
	"time"

	"github.com/pkg/errors"

	"github.com/andviro/grayproxy/pkg/gelf"
)

type Listener struct {
	Address                                                            string
	MaxChunkSize, MaxMessageSize, DecompressSizeLimit, AssembleTimeout int
}

func (l *Listener) address() string {
	if l.Address != "" {
		return l.Address
	}
	return ":12201"
}

func (in *Listener) Listen(dest chan<- gelf.Chunk) (err error) {
	chunks := make(chan gelf.Chunk)
	defer close(chunks)

	decodedMsgs := gelf.Extract(gelf.Assemble(chunks, in.MaxMessageSize, time.Millisecond*time.Duration(in.AssembleTimeout)), in.DecompressSizeLimit)
	go func() {
		for msg := range decodedMsgs {
			dest <- msg
		}
	}()

	l, err := net.ListenPacket("udp", in.address())
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
