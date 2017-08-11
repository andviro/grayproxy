package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	web "github.com/go-mixins/http"
	"github.com/pkg/errors"

	"github.com/andviro/grayproxy/gelf"
)

type httpListener struct {
	web.Server
}

func (in *httpListener) listen(dest chan gelf.Chunk) (err error) {
	err = in.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 400)
		}
		dest <- data
	}))
	return
}

func (in *httpListener) String() string {
	return fmt.Sprint(*in)
}

func (in *httpListener) Set(val string) error {
	n, _ := fmt.Sscan(strings.TrimPrefix(val, "http://"), &in.Address, &in.StopTimeout)
	if n == 0 {
		return errors.New("empty input description")
	}
	if in.StopTimeout == 0 {
		in.StopTimeout = 5000
	}
	return nil
}

type httpSender struct {
	Address     string
	SendTimeout int
}

func (out *httpSender) send(data []byte) (err error) {
	req, err := http.NewRequest("POST", out.Address, bytes.NewReader(data))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: time.Duration(out.SendTimeout) * time.Millisecond,
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		err = errors.New(resp.Status)
	}
	return
}

func (out *httpSender) String() string {
	return fmt.Sprint(*out)
}

func (out *httpSender) Set(val string) error {
	n, _ := fmt.Sscan(val, &out.Address, &out.SendTimeout)
	if n == 0 {
		return errors.New("empty input description")
	}
	if out.SendTimeout == 0 {
		out.SendTimeout = 500
	}
	return nil
}
