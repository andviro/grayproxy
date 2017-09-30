// Package http provides basic server skeleton that listens on specified
// http:// or unix:// socket and exits correctly on keyboard breaks and OS
// signals.
//
//
package http

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

// Server is a scaffolding for basic HTTP server. The structure members are
// configured externally, through JSON deserialization or environment variables
type Server struct {
	StopTimeout int    `default:"5000"`
	Address     string `default:":8000"`
}

// Serve starts the server on the configured address
func (srv *Server) Serve(h http.Handler) (err error) {
	l, err := srv.listen()
	if err != nil {
		return
	}
	serverErrors := make(chan error, 1)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGKILL)
	defer signal.Stop(interrupt)
	hs := http.Server{Handler: h}

	go func() {
		serverErrors <- hs.Serve(l)
	}()

	select {
	case sig := <-interrupt:
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(srv.StopTimeout)*time.Millisecond)
		defer cancel()
		if err = hs.Shutdown(ctx); err != nil {
			return
		}
		err = errors.New(sig.String())
	case err = <-serverErrors:
		break
	}
	return
}
