package main

import (
	"fmt"
	"strings"

	"github.com/andviro/grayproxy/gelf"
	"github.com/pkg/errors"
)

type input interface {
	listen(dest chan gelf.Chunk) (err error)
	String() string
	Set(val string) error
}

type inputs []input

func (ins *inputs) String() string {
	return fmt.Sprint(*ins)
}

func (ins *inputs) Set(val string) (err error) {
	vals := strings.Split(val, ",")
	*ins = make([]input, len(vals))
	for i, v := range vals {
		switch {
		case strings.HasPrefix(v, "udp://"):
			(*ins)[i] = new(udpListener)
		case strings.HasPrefix(v, "http://"):
			(*ins)[i] = new(httpListener)
		default:
			(*ins)[i] = new(tcpListener)
		}
		if err = (*ins)[i].Set(v); err != nil {
			return errors.Wrapf(err, "parsing input %d", i+1)
		}
	}
	return
}
