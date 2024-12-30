package cocurrent

import "sync"

//
// @Author yfy2001
// @Date 2024/12/20 09 38
//

type SafeMap[K comparable, V any] struct {
	rwm  sync.RWMutex
	data map[K]V
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		rwm:  sync.RWMutex{},
		data: make(map[K]V),
	}
}

func NewSafeMapFromMap[K comparable, V any](m map[K]V) *SafeMap[K, V] {
	sm := NewSafeMap[K, V]()
	for k, v := range m {
		sm.Set(k, v)
	}
	return sm
}

func (s *SafeMap[K, V]) Set(k K, v V) {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	s.data[k] = v
}

func (s *SafeMap[K, V]) Get(k K) (v V, ok bool) {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	v, ok = s.data[k]
	return v, ok
}

func (s *SafeMap[K, V]) Delete(k K) {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	delete(s.data, k)
}

func (s *SafeMap[K, V]) Clear() {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	s.data = make(map[K]V)
}
