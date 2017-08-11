package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
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
