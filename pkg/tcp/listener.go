package tcp

import (
	"bufio"
	"bytes"
	"net"

	"github.com/armon/go-proxyproto"
	"github.com/pkg/errors"

	"github.com/andviro/grayproxy/pkg/gelf"
)

type Listener struct {
	Address string
}

func tcpSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, 0); i >= 0 {
		buf := make([]byte, i)
		copy(buf, data[0:i])
		return i + 1, buf, nil
	}
	if atEOF {
		buf := make([]byte, len(data))
		copy(buf, data)
		return len(buf), buf, nil
	}
	return
}

func (l *Listener) Listen(dest chan<- gelf.Chunk) (err error) {
	lis, err := net.Listen("tcp", l.Address)
	if err != nil {
		return errors.Wrap(err, "setting up TCP listener")
	}
	lis = &proxyproto.Listener{Listener: lis}
	for {
		conn, err := lis.Accept()
		if err != nil {
			return errors.Wrap(err, "accepting connection")
		}
		scanner := bufio.NewScanner(conn)
		scanner.Split(tcpSplit)
		for scanner.Scan() {
			dest <- scanner.Bytes()
		}
		if err := scanner.Err(); err != nil {
			return errors.Wrap(err, "scanning input")
		}
	}
}
