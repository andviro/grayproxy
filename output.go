package main

import (
	"bytes"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"
)

// Output represents GELF TCP or HTTP output
type Output struct {
	Address     string
	SendTimeout int
}

func (out Output) sendHTTP(data []byte) (err error) {
	req, err := http.NewRequest("POST", out.Address+"/gelf", bytes.NewReader(data))
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

func (out Output) sendTCP(data []byte) (err error) {
	conn, err := net.DialTimeout("tcp", out.Address, time.Duration(out.SendTimeout)*time.Millisecond)
	if err != nil {
		return
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(time.Duration(out.SendTimeout) * time.Millisecond))
	n, err := conn.Write(data)
	if err != nil {
		return
	}
	if n != len(data) {
		return errors.New("short TCP write")
	}
	return
}

// Send sends message to required address
func (out Output) Send(data []byte) error {
	switch {
	case strings.HasPrefix(out.Address, "http"):
		return out.sendHTTP(data)
	}
	return out.sendTCP(data)
}
