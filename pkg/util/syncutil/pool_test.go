package syncutil

import (
	testingx "github.com/octohelm/x/testing"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	count := 0

	p := NewPool(func() (int, error) {
		count++
		time.Sleep(1 * time.Second)
		return count, nil
	})

	for i := 0; i < 10; i++ {
		go func() {
			v := p.MustGet()
			testingx.Expect(t, v, testingx.Be(1))
		}()
	}
}
