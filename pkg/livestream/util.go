package livestream

import "sync"

type XMap[K any, V any] struct {
	sync.Map
}

func (x *XMap[K, V]) Load(k K) (found V, ok bool) {
	got, ok := x.Map.Load(k)
	if ok {
		return got.(V), true
	}
	return
}
func (x *XMap[K, V]) Store(k K, v V) {
	x.Map.Store(k, v)
}

func (x *XMap[K, V]) LoadOrStore(k K, v V) (found V, ok bool) {
	got, ok := x.Map.LoadOrStore(k, v)
	if ok {
		return got.(V), true
	}
	return
}

func (x *XMap[K, V]) LoadAndDelete(k K) (found V, ok bool) {
	got, ok := x.Map.LoadAndDelete(k)
	if ok {
		return got.(V), true
	}
	return
}
