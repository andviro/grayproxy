package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/nsqio/go-diskqueue"
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
		return &udpListener{
			Address:             strings.TrimPrefix(addr, "udp://"),
			MaxChunkSize:        maxChunkSize,
			MaxMessageSize:      -1,
			DecompressSizeLimit: decompressSizeLimit,
			AssembleTimeout:     assembleTimeout,
		}
	case strings.HasPrefix(addr, "http://"):
		l := new(httpListener)
		l.Address = strings.TrimPrefix(addr, "http://")
		l.StopTimeout = stopTimeout
		return l
	}
	return &tcpListener{Address: strings.TrimPrefix(addr, "tcp://")}
}

func dummyLogf(lvl diskqueue.LogLevel, f string, args ...interface{}) {}

func (app *app) configure() (err error) {
	fs := flag.NewFlagSet("grayproxy", flag.ExitOnError)
	fs.Var(&app.inputURLs, "in", "input address in form schema://address:port (may be specified multiple times). Default: udp://:12201")
	fs.Var(&app.outputURLs, "out", "output address in form schema://address:port (may be specified multiple times)")
	fs.IntVar(&app.sendTimeout, "sendTimeout", 1000, "maximum TCP or HTTP output timeout (ms)")
	fs.StringVar(&app.dataDir, "dataDir", "", "buffer directory (defaults to no buffering)")
	if err = fs.Parse(os.Args[1:]); err != nil {
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
	app.outs = make([]sender, len(app.outputURLs))
	app.sendErrors = make([]error, len(app.outputURLs))
	for i, v := range app.outputURLs {
		switch {
		case strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://"):
			app.outs[i] = &httpSender{Address: v, SendTimeout: app.sendTimeout}
		default:
			app.outs[i] = &tcpSender{Address: strings.TrimPrefix(v, "tcp://"), SendTimeout: app.sendTimeout}
		}
		log.Printf("Added output %d: %s", i, v)
	}
	if app.dataDir == "" {
		app.q = newDummyQueue()
		log.Println("Buffering is not configured, unsent messages will be lost")
		return
	}

	stat, err := os.Stat(app.dataDir)
	if err != nil {
		return errors.Wrap(err, "checking buffer directory")
	}
	if !stat.IsDir() {
		return errors.Errorf("%q is not a directory", app.dataDir)
	}
	app.q = diskqueue.New("messages", app.dataDir, diskFileSize, 0, decompressSizeLimit, 2500, 2000, dummyLogf)
	return
}
