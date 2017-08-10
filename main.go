package main

import (
	"flag"
	"log"
	"strings"
	"sync"

	"github.com/andviro/grayproxy/gelf"
)

var inputs = flag.String("in", "udp://0.0.0.0:12221", "comma-separated list of GELF input addresses")
var outputs = flag.String("out", "tcp://127.0.0.0:12221", "comma-separated list of output Graylog servers (TCP or HTTP)")

func init() {
	flag.Parse()
}

func main() {
	msgs := make(chan gelf.Chunk, 100)
	defer close(msgs)

	oaddrs := strings.Split(*outputs, ",")
	outs := make([]Output, len(oaddrs))
	for i, addr := range oaddrs {
		outs[i].Address = addr
	}

	iaddrs := strings.Split(*inputs, ",")
	ins := make([]Input, len(iaddrs))
	for i, addr := range iaddrs {
		ins[i].Address = addr
	}

	go func() {
		for msg := range msgs {
			var sent bool
			for i := range outs {
				if err := outs[i].Send(msg); err == nil {
					sent = true
					break
				}
			}
			if !sent {
				log.Println("Message not sent")
			}
		}
	}()

	var wg sync.WaitGroup
	for i, in := range ins {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := in.Run(msgs)
			if err != nil {
				log.Printf("Input %d exited with error %+v", i, err)
			}
		}(i)
	}
	wg.Wait()
	log.Print("Bye")
}
