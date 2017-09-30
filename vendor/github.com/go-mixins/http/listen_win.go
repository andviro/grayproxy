// +build windows

package http

import (
	"net"
	"strings"
)

func (srv *Server) listen() (l net.Listener, err error) {
	switch addr := srv.Address; {
	case strings.HasPrefix(addr, "http://"):
		addr = strings.TrimPrefix(addr, "http://")
		fallthrough
	default:
		l, err = net.Listen("tcp", addr)
	}
	return
}
