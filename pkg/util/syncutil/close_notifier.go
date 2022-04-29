package syncutil

import "sync/atomic"

func NewCloseNotifier() CloseNotifier {
	return CloseNotifier{
		done: make(chan error, 1),
	}
}

type CloseNotifier struct {
	closed uint32
	done   chan error
}

func (c *CloseNotifier) Closed() bool {
	return atomic.LoadUint32(&c.closed) != 0
}

func (c *CloseNotifier) SendDone(err error) {
	if atomic.LoadUint32(&c.closed) == 0 {
		c.done <- err
	}
}

func (c *CloseNotifier) Done() <-chan error {
	return c.done
}

func (c *CloseNotifier) Close() error {
	if atomic.LoadUint32(&c.closed) == 0 {
		close(c.done)
		atomic.StoreUint32(&c.closed, 1)
	}
	return nil
}

func (c *CloseNotifier) Reset() {
	c.done = make(chan error, 1)
	atomic.StoreUint32(&c.closed, 0)
}
