package disk

import (
	"path/filepath"
	"sync"
	"time"

	ring "github.com/cloudflare/buffer"
	"github.com/pkg/errors"
)

// Queue implements on-disk buffering queue
type Queue struct {
	buf  *ring.Buffer
	r    chan []byte
	stop chan struct{}
	wg   sync.WaitGroup
}

func New(dataDir string, fileSize int) (*Queue, error) {
	buf, err := ring.New(filepath.Join(dataDir, "queue"), fileSize)
	if err != nil {
		return nil, errors.Wrap(err, "new buffer")
	}
	q := &Queue{buf: buf, r: make(chan []byte)}
	q.wg.Add(1)
	go func() {
		defer close(q.r)
		defer q.wg.Done()
		delay := 10 * time.Millisecond
		for {
			data, err := q.buf.Pop()
			if err != nil {
				return
			}
			if data == nil {
				time.Sleep(delay)
				select {
				case <-q.stop:
					return
				default:
					continue
				}
			}
			select {
			case <-q.stop:
				return
			case q.r <- data:
				break
			}
		}
	}()
	return q, nil
}

func (q *Queue) Put(data []byte) error {
	if err := q.buf.Insert(data); err != nil {
		return errors.Wrap(err, "put message to buffer")
	}
	return nil
}

func (q *Queue) ReadChan() <-chan []byte {
	return q.r
}

func (q *Queue) Close() error {
	close(q.stop)
	q.wg.Wait()
	return nil
}
