package loki

import (
	"log"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/grafana/loki/pkg/promtail/client"
	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
)

type Handler interface {
	Handle(ls model.LabelSet, t time.Time, s string) error
}

type Sender struct {
	Job     string
	Handler Handler
}

type zeroLog struct{}

func (zeroLog) Log(keyvals ...interface{}) error {
	log.Println(keyvals...)
	return nil
}

func New(addr string) (*Sender, error) {
	var cfg client.Config
	if err := cfg.URL.Set(addr); err != nil {
		return nil, errors.Wrap(err, "set client URL")
	}
	c, err := client.New(cfg, zeroLog{})
	if err != nil {
		return nil, errors.Wrap(err, "create client")
	}
	return &Sender{Handler: c}, nil
}

func (s *Sender) job() model.LabelValue {
	if s.Job == "" {
		return "grayproxy"
	}
	return model.LabelValue(s.Job)
}

func (s *Sender) labelSet() model.LabelSet {
	res := make(model.LabelSet)
	res["job"] = s.job()
	return res
}

func (s *Sender) Send(data []byte) (err error) {
	var msg string
	ls := s.labelSet()
	ts := time.Now()
	jsonparser.ObjectEach(data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		switch k := strings.TrimPrefix(string(key), "_"); k {
		case "timestamp":
			ms, _ := jsonparser.ParseFloat(value)
			ts = time.Unix(0, int64(ms*float64(time.Second)))
		case "short_message":
			if msg == "" {
				msg = string(value)
			}
		case "full_message":
			if string(value) != "" {
				msg = string(value)
			}
		default:
			ls[model.LabelName(k)] = model.LabelValue(string(value))
		}
		return nil
	})
	if ts.IsZero() {
		ts = time.Now()
	}
	return errors.Wrap(s.Handler.Handle(ls, ts, msg), "handle log entry")
}
