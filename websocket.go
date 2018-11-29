package main

import (
	"log"
	"net/http"

	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/jeremywohl/flatten"
	"net/url"
	"strings"
	"time"
)

var wsClients = make(map[*wsListener]bool)

type wsListener struct {
	c      *websocket.Conn
	filter map[string][]string
}

func (l *wsListener) close(msg string) {
	l.c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, msg))
	time.Sleep(2)
	l.c.Close()
}

type wsSender struct {
	Address string
	URL     *url.URL
}

func newSender(address string) sender {
	log.Print("Creating WS Sender")

	uri, err := url.Parse(address)
	if err != nil {
		panic(err)
	}

	s := wsSender{
		Address: address,
		URL:     uri,
	}
	go s.Listen()
	return &s
}

func (out *wsSender) logs(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	log.Printf("Connection from: [%s] [%s] [%v]", r.RemoteAddr, r.RequestURI, r.URL.Query())
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
	if out.URL.User.Username() != "" {
		if token, ok := query["token"]; ok {
			if token[0] != out.URL.User.Username() {
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

	wsClients[&l] = true

	log.Printf("filter: %v", l.filter)

	for {
		_, _, err := l.c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		// log.Printf("recv: %d %s", mt, message)
	}
	delete(wsClients, &l)
	log.Print("Disconnected: ", r.RemoteAddr)
}

func (out *wsSender) Listen() {
	http.HandleFunc("/", out.logs)
	log.Fatal(http.ListenAndServe(out.URL.Host, nil))
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
	// log.Printf("RAW %v", m.kv)

	var newKV = make(map[string]interface{})
	for k, v := range m.kv {
		newKV[strings.TrimPrefix(k, "_")] = v
	}
	m.kv = newKV

	// log.Printf("RAW FIXED %v", m.kv)

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
				// log.Printf("NOT SENDING: [filter_k:%s] [filter_vs:%v] [val:%s]", filter_k, filter_vs, val)
				return false
			}
			// log.Printf("FOUND [%s] = [%s] ; %v", k, val, vs)
		} else {
			// log.Printf("NOT SENDING: [filter_k:%s] [filter_vs:%v]", filter_k, filter_vs)
			return false
		}
	}
	return true
}

func (out *wsSender) Send(data []byte) (err error) {

	m, err := newMsg(data)
	if err != nil {
		return
	}
	// log.Print(m.flat)

	if len(wsClients) > 0 {
		for l, _ := range wsClients {
			if m.isSendable(l.filter) {
				log.Print("wsSender sending: ", string(m.json))
				l.c.WriteMessage(1, m.json)
			}
		}
	}
	return
}
