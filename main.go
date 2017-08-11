package main

import (
	"flag"
	"log"
	"sync"

	"github.com/andviro/grayproxy/gelf"
)

var ins inputs
var outs outputs

func init() {
	flag.Var(&ins, "in", "comma-separated list of GELF inputs")
	flag.Var(&outs, "out", "comma-separated list of outputs")
}

func main() {
	flag.Parse()
	if len(ins) == 0 {
		log.Fatal("no inputs specified")
	}
	if len(outs) == 0 {
		log.Fatal("no outputs specified")
	}

	msgs := make(chan gelf.Chunk, 100)
	defer close(msgs)
	go func() {
		for msg := range msgs {
			var sent bool
			for i := range outs {
				if err := outs[i].send(msg); err == nil {
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
			err := in.listen(msgs)
			if err != nil {
				log.Printf("Input %d exited with error %+v", i, err)
			}
		}(i)
	}
	wg.Wait()
	log.Print("Bye")
}
