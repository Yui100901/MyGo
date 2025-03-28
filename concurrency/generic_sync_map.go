package concurrency

import (
	"sync"
)

//
// @Author yfy2001
// @Date 2024/12/30 10 41
//

// GenericSyncMap 是一个使用 sync.Map 封装的泛型 Map
type GenericSyncMap[K comparable, V any] struct {
	data sync.Map
}

func NewGenericSyncMap[K comparable, V any]() *GenericSyncMap[K, V] {
	return &GenericSyncMap[K, V]{
		data: sync.Map{},
	}
}

func NewGenericSyncMapFromMap[K comparable, V any](m map[K]V) *GenericSyncMap[K, V] {
	gm := NewGenericSyncMap[K, V]()
	for k, v := range m {
		gm.Set(k, v)
	}
	return gm
}

// Set 方法将键值对添加到 Map 中
func (gm *GenericSyncMap[K, V]) Set(key K, value V) {
	gm.data.Store(key, value)
}

// Get 方法从 Map 中检索值
func (gm *GenericSyncMap[K, V]) Get(key K) (V, bool) {
	value, ok := gm.data.Load(key)
	if ok {
		return value.(V), true
	}
	var zeroValue V
	return zeroValue, false
}

func (gm *GenericSyncMap[K, V]) GetOr(k K, dV V) (v V) {
	value, ok := gm.data.Load(k)
	if ok {
		return value.(V)
	}
	return dV
}

func (gm *GenericSyncMap[K, V]) GetOrElse(k K, fn func() V) (v V) {
	value, ok := gm.data.Load(k)
	if ok {
		return value.(V)
	}
	return fn()
}

// Delete 方法从 Map 中删除键值对
func (gm *GenericSyncMap[K, V]) Delete(key K) {
	gm.data.Delete(key)
}

// Keys 方法返回 Map 中的所有键
func (gm *GenericSyncMap[K, V]) Keys() []K {
	var keys []K
	gm.data.Range(func(key, value any) bool {
		keys = append(keys, key.(K))
		return true
	})
	return keys
}

// Values 方法返回 Map 中的所有值
func (gm *GenericSyncMap[K, V]) Values() []V {
	var values []V
	gm.data.Range(func(key, value any) bool {
		values = append(values, value.(V))
		return true
	})
	return values
}

// Range 方法遍历 Map 中的所有键值对
func (gm *GenericSyncMap[K, V]) Range(f func(key K, value V) bool) {
	gm.data.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}
