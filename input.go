package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/andviro/grayproxy/gelf"
	"github.com/armon/go-proxyproto"
	web "github.com/go-mixins/http"
)

// Input represents generic TCP, UDP or HTTP GELF input
type Input struct {
	Address                                                            string `default:"udp://0.0.0.0:12221"`
	MaxChunkSize, MaxMessageSize, DecompressSizeLimit, AssembleTimeout int
}

func (in Input) listenHTTP(dest chan gelf.Chunk) (err error) {
	defer close(dest)
	srv := &web.Server{Address: in.Address, StopTimeout: 1000}
	err = srv.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 400)
		}
		dest <- data
	}))
	return
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

func (in Input) listenTCP(dest chan gelf.Chunk) (err error) {
	defer close(dest)
	l, err := net.Listen("tcp", strings.TrimPrefix(in.Address, "tcp://"))
	if err != nil {
		return
	}
	l = &proxyproto.Listener{Listener: l}
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(conn)
		scanner.Split(tcpSplit)
		for scanner.Scan() {
			dest <- scanner.Bytes()
		}
		if err := scanner.Err(); err != nil {
			return err
		}
	}
}

func (in Input) listenUDP(dest chan gelf.Chunk) (err error) {
	defer close(dest)
	l, err := net.ListenPacket("udp", strings.TrimPrefix(in.Address, "udp://"))
	if err != nil {
		return
	}
	buf := make([]byte, in.MaxChunkSize)
	for {
		n, _, err := l.ReadFrom(buf)
		if err != nil {
			return err
		}
		dest <- buf[:n]
	}
}

// Run starts listening on input. The received messages are passed to the
// output channel
func (in Input) Run(output chan gelf.Chunk) error {
	chunks := make(chan gelf.Chunk)
	decodedMsgs := gelf.Extract(gelf.Assemble(chunks, in.MaxMessageSize, time.Millisecond*time.Duration(in.AssembleTimeout)), in.DecompressSizeLimit)
	go func() {
		for msg := range decodedMsgs {
			output <- msg
		}
	}()
	switch {
	case strings.HasPrefix(in.Address, "http://"):
		return in.listenHTTP(chunks)
	case strings.HasPrefix(in.Address, "tcp://"):
		return in.listenTCP(chunks)
	default:
		return in.listenUDP(chunks)
	}
}
