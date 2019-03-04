package tcp

import (
	"net"
	"time"

	"github.com/pkg/errors"
)

type Sender struct {
	Address     string
	SendTimeout int
	conn        net.Conn
	err         error
}

func (s *Sender) write(data []byte) {
	if s.err != nil {
		return
	}
	defer func() {
		if s.err != nil && s.conn != nil {
			s.conn.Close()
			s.conn = nil
		}
	}()
	var n int
	if n, s.err = s.conn.Write(data); s.err != nil {
		s.err = errors.Wrap(s.err, "writing TCP")
		return
	}
	if n != len(data) {
		s.err = errors.New("short TCP write")
	}
}

func (s *Sender) Send(data []byte) (err error) {
	if s.conn == nil {
		s.err = nil
		if s.conn, err = net.DialTimeout("tcp", s.Address, time.Duration(s.SendTimeout)*time.Millisecond); err != nil {
			s.err = errors.Wrap(err, "creating TCP connection")
			return s.err
		}
	}
	s.conn.SetDeadline(time.Now().Add(time.Duration(s.SendTimeout) * time.Millisecond))
	s.write(data)
	s.write([]byte{0})
	return s.err
}
