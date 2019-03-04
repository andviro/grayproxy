package loki_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/andviro/goldie"
	"github.com/andviro/grayproxy/pkg/loki"
	"github.com/prometheus/common/model"
)

type testHandler bytes.Buffer

func (t *testHandler) Handle(ls model.LabelSet, ts time.Time, s string) error {
	buf := (*bytes.Buffer)(t)
	jd, _ := json.Marshal(ts)
	fmt.Fprintf(buf, "%s\n", jd)
	jd, _ = json.MarshalIndent(ls, "", "\t")
	fmt.Fprintf(buf, "%s\n", jd)
	fmt.Fprintf(buf, "%s\n", s)
	return nil
}

func TestSender_Send(t *testing.T) {
	buf := new(bytes.Buffer)
	s := loki.Sender{Handler: (*testHandler)(buf), Job: "test"}
	err := s.Send([]byte(`{
	  "version": "1.1",
	  "host": "example.org",
	  "short_message": "A short message that helps you identify what is going on",
	  "full_message": "Backtrace here\n\nmore stuff",
	  "timestamp": 1385053862.3072,
	  "level": 1,
	  "_user_id": 9001,
	  "_some_info": "foo",
	  "_some_env_var": "bar"
	}`))
	if err != nil {
		t.Fatal(err)
	}
	goldie.Assert(t, "sender-send", buf.Bytes())
}
