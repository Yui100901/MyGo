package cocurrent

import "sync"

//
// @Author yfy2001
// @Date 2024/12/20 09 38
//

// SafeMap 使用读写锁的线程安全map
type SafeMap[K comparable, V any] struct {
	rwm  sync.RWMutex
	data map[K]V
}

// NewSafeMap 创建一个线程安全的map
func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		rwm:  sync.RWMutex{},
		data: make(map[K]V),
	}
}

// NewSafeMapFromMap 将普通map转换为SafeMap
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

func (s *SafeMap[K, V]) GetOr(k K, dV V) (v V) {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	v, ok := s.data[k]
	if ok {
		return v
	}
	return dV
}

func (s *SafeMap[K, V]) GetOrElse(k K, fn func() V) (v V) {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	v, ok := s.data[k]
	if ok {
		return v
	}
	return fn()
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

// Keys 返回 SafeMap 中的所有键
func (s *SafeMap[K, V]) Keys() []K {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	keys := make([]K, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

// Values 返回 SafeMap 中的所有值
func (s *SafeMap[K, V]) Values() []V {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	values := make([]V, 0, len(s.data))
	for _, v := range s.data {
		values = append(values, v)
	}
	return values
}

func (s *SafeMap[K, V]) ForEach(fn func(K, V) bool) {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	for k, v := range s.data {
		s.rwm.RUnlock()
		fn(k, v)
		s.rwm.RLock()
	}
}
