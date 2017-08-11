package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type output interface {
	send(data []byte) (err error)
	String() string
	Set(val string) error
}

type outputs []output

func (outs *outputs) String() string {
	return fmt.Sprint(*outs)
}

func (outs *outputs) Set(val string) (err error) {
	vals := strings.Split(val, ",")
	*outs = make([]output, len(vals))
	for i, v := range vals {
		switch {
		case strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://"):
			(*outs)[i] = new(httpSender)
		default:
			(*outs)[i] = new(tcpSender)
		}
		if err = (*outs)[i].Set(v); err != nil {
			return errors.Wrapf(err, "parsing input %d", i+1)
		}
	}
	return
}
