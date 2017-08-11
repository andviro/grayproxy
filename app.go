package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/andviro/grayproxy/gelf"
)

type urlList []string

func (ul *urlList) Set(val string) error {
	*ul = append(*ul, val)
	return nil
}

func (ul *urlList) String() string {
	return strings.Join(*ul, ",")
}

type listener interface {
	listen(dest chan gelf.Chunk) (err error)
}

type sender interface {
	send(data []byte) (err error)
}

type app struct {
	inputURLs                                         urlList
	outputURLs                                        urlList
	maxChunkSize, maxMessageSize, decompressSizeLimit int
	assembleTimeout, stopTimeout, sendTimeout         int

	ins  []listener
	outs []sender
}

func (app *app) configure() (err error) {
	fs := flag.NewFlagSet("grayproxy", flag.ExitOnError)
	fs.Var(&app.inputURLs, "in", "input address in form schema://address:port (may be specified multiple times). Default: udp://:12201")
	fs.Var(&app.outputURLs, "out", "output address in form schema://address:port (may be specified multiple times)")
	fs.IntVar(&app.maxChunkSize, "maxChunkSize", 8192, "maximum UDP chunk size")
	fs.IntVar(&app.maxMessageSize, "maxMessageSize", 128*1024, "maximum UDP de-chunked message size")
	fs.IntVar(&app.decompressSizeLimit, "decompressSizeLimit", 1024*1024, "maximum decompressed message size")
	fs.IntVar(&app.assembleTimeout, "assembleTimeout", 1000, "maximum UDP chunk assemble time (ms)")
	fs.IntVar(&app.sendTimeout, "sendTimeout", 1000, "maximum TCP or HTTP output timeout (ms)")
	fs.IntVar(&app.stopTimeout, "stopTimeout", 2000, "server stop timeout (ms)")
	if err = fs.Parse(os.Args[1:]); err != nil {
		return errors.Wrap(err, "parsing command-line")
	}
	if len(app.inputURLs) == 0 {
		app.inputURLs = urlList{"udp://:12201"}
	}
	app.ins = make([]listener, len(app.inputURLs))
	for i, v := range app.inputURLs {
		switch {
		case strings.HasPrefix(v, "udp://"):
			app.ins[i] = &udpListener{
				Address:             strings.TrimPrefix(v, "udp://"),
				MaxChunkSize:        app.maxChunkSize,
				MaxMessageSize:      app.maxMessageSize,
				DecompressSizeLimit: app.decompressSizeLimit,
				AssembleTimeout:     app.assembleTimeout,
			}
		case strings.HasPrefix(v, "http://"):
			l := new(httpListener)
			l.Address = strings.TrimPrefix(v, "http://")
			l.StopTimeout = app.stopTimeout
			app.ins[i] = l
		default:
			app.ins[i] = &tcpListener{
				Address: strings.TrimPrefix(v, "tcp://"),
			}
		}
		log.Printf("Added input %d at %s", i, v)
	}

	if len(app.outputURLs) == 0 {
		log.Print("WARNING: no outputs configured")
	}
	app.outs = make([]sender, len(app.outputURLs))
	for i, v := range app.outputURLs {
		switch {
		case strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://"):
			app.outs[i] = &httpSender{Address: v, SendTimeout: app.sendTimeout}
		default:
			app.outs[i] = &tcpSender{Address: strings.TrimPrefix("tcp://", v), SendTimeout: app.sendTimeout}
		}
		log.Printf("Added output %d: %s", i, v)
	}
	return
}

func (app *app) run() (err error) {
	if err = app.configure(); err != nil {
		return errors.Wrap(err, "configuring app")
	}

	msgs := make(chan gelf.Chunk, len(app.ins))
	defer close(msgs)
	go func() {
		for msg := range msgs {
			var sent bool
			for i, out := range app.outs {
				if err := out.send(msg); err != nil {
					log.Printf("ERROR: sending message to output %d: %v", i, err)
					continue
				}
				sent = true
				break
			}
			if !sent {
				// TODO: message buffering on disk
				log.Printf("WARNING: message not sent: %q", string(msg))
			}
		}
	}()

	var wg sync.WaitGroup
	for i := range app.ins {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := app.ins[i].listen(msgs)
			if err != nil {
				log.Printf("Input %d exited with error: %+v", i, err)
			}
		}(i)
	}
	wg.Wait()
	log.Print("Bye")
	return
}
