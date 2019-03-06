package dummy

import (
	"errors"
	"sync"
	"time"
)

var Timeout = 10 * time.Second

type Queue struct {
	in   chan []byte
	out  chan []byte
	once sync.Once
}

func New() *Queue {
	q := new(Queue)
	q.in = make(chan []byte, 1)
	q.out = make(chan []byte)
	go func() {
		defer close(q.out)
		for msg := range q.in {
			q.out <- msg
		}
	}()
	return q
}

func (q *Queue) Put(data []byte) error {
	select {
	case q.in <- data:
		break
	case <-time.After(Timeout):
		return errors.New("overflow")
	}
	return nil
}

func (q *Queue) ReadChan() <-chan []byte {
	return q.out
}

func (q *Queue) Close() error {
	close(q.in)
	return nil
}
