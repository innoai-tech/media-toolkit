package livestream

import (
	"context"
	"sync/atomic"
)

type Switch struct {
	disabled int32
}

func (s *Switch) Enabled() bool {
	return atomic.LoadInt32(&s.disabled) == 0
}

func (s *Switch) Enable(ctx context.Context) {
	atomic.SwapInt32(&s.disabled, 0)
}

func (s Switch) Disable(ctx context.Context) {
	atomic.SwapInt32(&s.disabled, 1)
}
