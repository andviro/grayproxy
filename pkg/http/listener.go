package http

import (
	"io/ioutil"
	"net/http"

	web "github.com/go-mixins/http"

	"github.com/andviro/grayproxy/pkg/gelf"
)

type Listener struct {
	web.Server
}

func (l *Listener) Listen(dest chan<- gelf.Chunk) (err error) {
	err = l.Serve(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 400)
		}
		dest <- data
	}))
	return
}
