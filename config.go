package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/andviro/grayproxy/pkg/disk"
	"github.com/andviro/grayproxy/pkg/dummy"
	"github.com/andviro/grayproxy/pkg/http"
	"github.com/andviro/grayproxy/pkg/loki"
	"github.com/andviro/grayproxy/pkg/tcp"
	"github.com/andviro/grayproxy/pkg/udp"
	"github.com/andviro/grayproxy/pkg/ws"
	"github.com/pkg/errors"
)

const (
	maxChunkSize        = 8192
	assembleTimeout     = 1000
	stopTimeout         = 2000
	decompressSizeLimit = 1048576
	diskFileSize        = 104857600
)

type urlList []string

func (ul *urlList) Set(val string) error {
	*ul = append(*ul, val)
	return nil
}

func (ul *urlList) String() string {
	return strings.Join(*ul, ",")
}

func (app *app) newListener(addr string) listener {
	switch {
	case strings.HasPrefix(addr, "udp://"):
		return &udp.Listener{
			Address:             strings.TrimPrefix(addr, "udp://"),
			MaxChunkSize:        maxChunkSize,
			MaxMessageSize:      -1,
			DecompressSizeLimit: decompressSizeLimit,
			AssembleTimeout:     assembleTimeout,
		}
	case strings.HasPrefix(addr, "http://"):
		l := new(http.Listener)
		l.Address = strings.TrimPrefix(addr, "http://")
		l.StopTimeout = stopTimeout
		return l
	}
	return &tcp.Listener{Address: strings.TrimPrefix(addr, "tcp://")}
}

func (app *app) configure() error {
	fs := flag.NewFlagSet("grayproxy", flag.ExitOnError)
	fs.Var(&app.inputURLs, "in", "input address in form schema://address:port (may be specified multiple times). Default: udp://:12201")
	fs.Var(&app.outputURLs, "out", "output address in form schema://address:port (may be specified multiple times)")
	fs.BoolVar(&app.verbose, "v", false, "echo received logs on console")
	fs.IntVar(&app.sendTimeout, "sendTimeout", 1000, "maximum TCP or HTTP output timeout (ms)")
	fs.StringVar(&app.dataDir, "dataDir", "", "buffer directory (defaults to no buffering)")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return errors.Wrap(err, "parsing command-line")
	}
	if len(app.inputURLs) == 0 {
		app.inputURLs = urlList{"udp://:12201"}
	}
	app.ins = make([]listener, len(app.inputURLs))
	for i, v := range app.inputURLs {
		app.ins[i] = app.newListener(v)
		log.Printf("Added input %d at %s", i, v)
	}
	if len(app.outputURLs) == 0 {
		log.Print("WARNING: no outputs configured")
	}
	app.outs = make([]sender, 0, len(app.outputURLs))
	app.sendErrors = make([]error, len(app.outputURLs))
	for i, v := range app.outputURLs {
		log.Printf("adding output %d: %s", i, v)
		switch {
		case strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://"):
			if strings.HasSuffix(v, "/api/prom/push") {
				ls, err := loki.New(v)
				if err != nil {
					log.Println("could not create loki output: ", err.Error())
					continue
				}
				app.outs = append(app.outs, ls)
				break
			}
			app.outs = append(app.outs, &http.Sender{Address: v, SendTimeout: app.sendTimeout})
		case strings.HasPrefix(v, "ws://"):
			wss := &ws.Sender{Address: v}
			if err := wss.Start(); err != nil {
				log.Println("Invalid websocket URL: ", err.Error())
				break
			}
			app.outs = append(app.outs, wss)
		case strings.HasPrefix(v, "udp://"):
			app.outs = append(app.outs, &udp.Sender{Address: strings.TrimPrefix(v, "udp://"), SendTimeout: app.sendTimeout})
		default:
			app.outs = append(app.outs, &tcp.Sender{Address: strings.TrimPrefix(v, "tcp://"), SendTimeout: app.sendTimeout})
		}
	}
	if app.dataDir == "" {
		app.q = dummy.New()
		log.Println("Buffering is not configured, unsent messages will be lost")
		return nil
	}

	stat, err := os.Stat(app.dataDir)
	if err != nil {
		return errors.Wrap(err, "checking buffer directory")
	}
	if !stat.IsDir() {
		return errors.Errorf("%q is not a directory", app.dataDir)
	}
	q, err := disk.New(app.dataDir, diskFileSize)
	app.q = q
	return err
}
