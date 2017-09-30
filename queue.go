package main

type dummyQueue struct {
	in  chan []byte
	out chan []byte
}

func (dq *dummyQueue) Put(data []byte) error {
	dq.in <- data
	return nil
}

func (dq *dummyQueue) ReadChan() chan []byte {
	return dq.out
}

func (dq *dummyQueue) Close() error {
	close(dq.in)
	return nil
}

func newDummyQueue() (dq *dummyQueue) {
	dq = &dummyQueue{
		in:  make(chan []byte, 1),
		out: make(chan []byte),
	}
	go func() {
		defer close(dq.out)
		for msg := range dq.in {
			dq.out <- msg
		}
	}()
	return dq
}
