package main

import (
	"log"
	"sync"

	"github.com/andviro/grayproxy/pkg/gelf"
)

type listener interface {
	Listen(dest chan<- gelf.Chunk) (err error)
}

type sender interface {
	Send(data []byte) (err error)
}

type queue interface {
	Put(data []byte) error
	ReadChan() <-chan []byte
	Close() error
}

type app struct {
	inputURLs   urlList
	outputURLs  urlList
	sendTimeout int
	dataDir     string

	ins        []listener
	outs       []sender
	sendErrors []error
	q          queue
}

func (app *app) enqueue(msgs <-chan gelf.Chunk) {
	for msg := range msgs {
		if err := app.q.Put(msg); err != nil {
			panic(err)
		}
	}
}

func (app *app) dequeue() {
	for msg := range app.q.ReadChan() {
		var sent bool
		for i, out := range app.outs {
			err := out.Send(msg)
			if err != nil {
				if app.sendErrors[i] == nil {
					log.Printf("out %d: %v", i, err)
					app.sendErrors[i] = err
				}
				continue
			}
			if app.sendErrors[i] != nil {
				log.Printf("out %d is now alive", i)
				app.sendErrors[i] = nil
			}
			sent = true
			break
		}
		if !sent {
			if app.dataDir == "" {
				log.Println(string(msg))
				continue
			}
			if err := app.q.Put(msg); err != nil {
				panic(err)
			}
		}
	}
}

func (app *app) run() (err error) {
	if err = app.configure(); err != nil {
		return
	}
	defer app.q.Close()
	msgs := make(chan gelf.Chunk, len(app.ins)*1000000)
	defer close(msgs)
	var wg sync.WaitGroup
	for i := range app.ins {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := app.ins[i].Listen(msgs)
			if err != nil {
				log.Printf("Input %d exited with error: %+v", i, err)
			}
		}(i)
	}
	go app.enqueue(msgs)
	go app.dequeue()
	wg.Wait()
	return
}
