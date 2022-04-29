package syncutil

import (
	"sync/atomic"
)

type Once struct {
	done uint32
}

func (o *Once) Do(f func() error) error {
	if atomic.LoadUint32(&o.done) == 0 {
		if err := f(); err != nil {
			return err
		}
		atomic.StoreUint32(&o.done, 1)
	}
	return nil
}

func (o *Once) Ready() bool {
	return atomic.LoadUint32(&o.done) != 0
}

func (o *Once) Reset() {
	atomic.StoreUint32(&o.done, 0)
}
