package main

import (
	"bufio"
	"bytes"
	"net"
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
	l, err := net.Listen("tcp", in.Address)
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
	conn        net.Conn
	err         error
}

func (out *tcpSender) write(data []byte) {
	if out.err != nil {
		return
	}
	var n int
	if n, out.err = out.conn.Write(data); out.err != nil {
		out.err = errors.Wrap(out.err, "writing TCP")
		return
	}
	if n != len(data) {
		out.err = errors.New("short TCP write")
	}
}

func (out *tcpSender) send(data []byte) (err error) {
	if out.conn, err = net.DialTimeout("tcp", out.Address, time.Duration(out.SendTimeout)*time.Millisecond); err != nil {
		return errors.Wrap(err, "creating TCP connection")
	}
	defer out.conn.Close()
	out.conn.SetDeadline(time.Now().Add(time.Duration(out.SendTimeout) * time.Millisecond))
	out.write(data)
	out.write([]byte{0})
	return out.err
}
