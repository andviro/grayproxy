package http

import (
	"bytes"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type Sender struct {
	Address     string
	SendTimeout int
}

func (s *Sender) Send(data []byte) (err error) {
	req, err := http.NewRequest("POST", s.Address, bytes.NewReader(data))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: time.Duration(s.SendTimeout) * time.Millisecond,
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
