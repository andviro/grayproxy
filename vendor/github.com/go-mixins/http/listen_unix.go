// +build !windows

package http

import (
	"net"
	"os"
	"strings"
	"syscall"
)

func (srv *Server) listen() (l net.Listener, err error) {
	switch addr := srv.Address; {
	case strings.HasPrefix(addr, "unix://"):
		addr = strings.TrimPrefix(addr, "unix://")
		if err = os.Remove(addr); err != nil {
			if !os.IsNotExist(err) {
				return
			}
			err = nil
		}
		oldmask := syscall.Umask(000)
		defer syscall.Umask(oldmask)
		l, err = net.Listen("unix", addr)
	case strings.HasPrefix(addr, "http://"):
		addr = strings.TrimPrefix(addr, "http://")
		fallthrough
	default:
		l, err = net.Listen("tcp", addr)
	}
	return
}
