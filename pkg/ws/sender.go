package main

import (
	"log"
	"net/http"
	"sync"

	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jeremywohl/flatten"
	"github.com/pkg/errors"
)

type wsListener struct {
	c      *websocket.Conn
	filter map[string][]string
}

func (l *wsListener) close(msg string) {
	l.c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, msg))
	time.Sleep(2)
	l.c.Close()
}

type Sender struct {
	Address string

	url     *url.URL
	once    sync.Once
	clients map[*wsListener]bool
}

func (s *Sender) Start() error {
	var rErr error
	s.once.Do(func() {
		u, err := url.Parse(s.Address)
		if err != nil {
			rErr = errors.Wrap(err, "parse URL")
			return
		}
		s.url = u
		s.clients = make(map[*wsListener]bool)
		go s.listenAndServe()
	})
	return rErr
}

func (s *Sender) logs(w http.ResponseWriter, r *http.Request) {
	s.once.Do(func() {
	})
	query := r.URL.Query()

	log.Printf("Connection from: [%s] [%s]", r.RemoteAddr, r.RequestURI)
	upgrader := websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	l := wsListener{
		c: c,
	}
	if s.url.User.Username() != "" {
		if token, ok := query["token"]; ok {
			if token[0] != s.url.User.Username() {
				log.Print("Unauthorized: ", r.RemoteAddr)
				l.close("Unauthorized")
				return
			}
		} else {
			log.Print("Unauthorized: ", r.RemoteAddr)
			l.close("Unauthorized")
			return
		}
	}

	delete(query, "token")

	l.filter = query

	s.clients[&l] = true

	for {
		_, _, err := l.c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
	}
	delete(s.clients, &l)
	log.Print("Disconnected: ", r.RemoteAddr)
}

func (s *Sender) listenAndServe() {
	http.HandleFunc("/", s.logs)
	log.Fatal(http.ListenAndServe(s.url.Host, nil))
}

type msg struct {
	raw  []byte
	flat string

	kv map[string]interface{}

	payload []interface{}

	json []byte
}

func newMsg(data []byte) (*msg, error) {
	var err error

	m := &msg{
		raw: data,
	}

	m.flat, err = flatten.FlattenString(string(data), "", flatten.DotStyle)
	if err != nil {
		return m, err
	}

	err = json.Unmarshal([]byte(m.flat), &m.kv)
	if err != nil {
		log.Printf("Error parsing JSON: %v", err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("Syntax error at byte offset %d", e.Offset)
		}
		return m, err
	}
	var newKV = make(map[string]interface{})
	for k, v := range m.kv {
		newKV[strings.TrimPrefix(k, "_")] = v
	}
	m.kv = newKV

	m.payload = append(m.payload, `incoming.gelf`)
	m.payload = append(m.payload, m.kv)

	m.json, err = json.Marshal(m.payload)
	if err != nil {
		log.Printf("Error parsing JSON: %v", err)
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("Syntax error at byte offset %d", e.Offset)
		}
		return m, err
	}

	return m, err
}

func (m *msg) isSendable(filter map[string][]string) bool {
	for filter_k, filter_vs := range filter {
		if val, ok := m.kv[filter_k]; ok {
			if len(filter_vs) == 0 || val != filter_vs[0] {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func (s *Sender) Send(data []byte) (err error) {

	m, err := newMsg(data)
	if err != nil {
		return
	}

	if len(s.clients) > 0 {
		for l, _ := range s.clients {
			if m.isSendable(l.filter) {
				// log.Print("Sender sending: ", string(m.json))
				l.c.WriteMessage(1, m.json)
			}
		}
	}
	return
}
