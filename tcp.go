package main

import (
	"bufio"
	"bytes"
	"net"
	"strings"
	"time"

	"github.com/armon/go-proxyproto"
	"github.com/pkg/errors"

	"github.com/andviro/grayproxy/gelf"
)

type tcpListener struct {
	Address string
}

func tcpSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, 0); i >= 0 {
		return i + 1, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return
}

func (in *tcpListener) listen(dest chan gelf.Chunk) (err error) {
	defer close(dest)
	l, err := net.Listen("tcp", strings.TrimPrefix(in.Address, "tcp://"))
	if err != nil {
		return errors.Wrap(err, "setting up TCP listener")
	}
	l = &proxyproto.Listener{Listener: l}
	for {
		conn, err := l.Accept()
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

type tcpSender struct {
	Address     string
	SendTimeout int
}

func (out *tcpSender) send(data []byte) (err error) {
	conn, err := net.DialTimeout("tcp", out.Address, time.Duration(out.SendTimeout)*time.Millisecond)
	if err != nil {
		return
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(time.Duration(out.SendTimeout) * time.Millisecond))
	n, err := conn.Write(data)
	if err != nil {
		return
	}
	if n != len(data) {
		return errors.New("short TCP write")
	}
	return
}
