package concurrency

import (
	"fmt"
	"sync"
)

//
// @Author yfy2001
// @Date 2024/12/20 09 38
//

const defaultShardCount = 32

// SafeMap 使用分片锁的线程安全map
type SafeMap[K comparable, V any] struct {
	shardCount uint64         // 分片数量
	locks      []sync.RWMutex // 读写锁数组，每个分片一个锁
	maps       []map[K]V      // 存储数据的分片数组
}

// NewSafeMap 创建一个带有指定分片数量的线程安全map
func NewSafeMap[K comparable, V any](shardCount int) *SafeMap[K, V] {
	if shardCount <= 0 {
		shardCount = defaultShardCount
	}

	m := &SafeMap[K, V]{
		shardCount: uint64(shardCount),
		locks:      make([]sync.RWMutex, shardCount),
		maps:       make([]map[K]V, shardCount),
	}

	for i := range m.maps {
		m.maps[i] = make(map[K]V)
	}

	return m
}

// NewSafeMapFromMap 将普通map转换为ConcurrentMap
func NewSafeMapFromMap[K comparable, V any](m map[K]V, shardCount int) *SafeMap[K, V] {
	safeMap := NewSafeMap[K, V](shardCount)
	for k, v := range m {
		safeMap.Set(k, v)
	}
	return safeMap
}

// Set 设置键值对
func (m *SafeMap[K, V]) Set(key K, value V) {
	shard := m.getShard(key)
	m.locks[shard].Lock()
	defer m.locks[shard].Unlock()
	m.maps[shard][key] = value
}

// Get 获取键对应的值，不存在时返回nil
func (m *SafeMap[K, V]) Get(key K) (V, bool) {
	shard := m.getShard(key)
	m.locks[shard].RLock()
	defer m.locks[shard].RUnlock()
	value, ok := m.maps[shard][key]
	return value, ok
}

// GetOr 获取键对应的值，不存在时返回给定的默认值
func (m *SafeMap[K, V]) GetOr(key K, dV V) V {
	shard := m.getShard(key)
	m.locks[shard].RLock()
	defer m.locks[shard].RUnlock()
	value, ok := m.maps[shard][key]
	if ok {
		return value
	}
	return dV
}

// GetOrElse 获取键对应的值，不存在时调用给定的函数并返回其结果
func (m *SafeMap[K, V]) GetOrElse(key K, fn func() V) V {
	shard := m.getShard(key)
	m.locks[shard].RLock()
	defer m.locks[shard].RUnlock()
	value, ok := m.maps[shard][key]
	if ok {
		return value
	}
	return fn()
}

// Delete 删除键对应的值
func (m *SafeMap[K, V]) Delete(key K) {
	shard := m.getShard(key)
	m.locks[shard].Lock()
	defer m.locks[shard].Unlock()
	delete(m.maps[shard], key)
}

// Clear 清空map
func (m *SafeMap[K, V]) Clear() {
	for shard := range m.maps {
		m.locks[shard].Lock()
		m.maps[shard] = make(map[K]V)
		m.locks[shard].Unlock()
	}
}

// Length 返回map中的元素数量
func (m *SafeMap[K, V]) Length() int {
	length := 0
	for shard := range m.maps {
		m.locks[shard].RLock()
		length += len(m.maps[shard])
		m.locks[shard].RUnlock()
	}
	return length
}

// Keys 返回map中的所有键
func (m *SafeMap[K, V]) Keys() []K {
	keys := make([]K, 0)
	for shard := range m.maps {
		m.locks[shard].RLock()
		for k := range m.maps[shard] {
			keys = append(keys, k)
		}
		m.locks[shard].RUnlock()
	}
	return keys
}

// Values 返回map中的所有值
func (m *SafeMap[K, V]) Values() []V {
	values := make([]V, 0)
	for shard := range m.maps {
		m.locks[shard].RLock()
		for _, v := range m.maps[shard] {
			values = append(values, v)
		}
		m.locks[shard].RUnlock()
	}
	return values
}

// ForEach 遍历map中的所有键值对
func (m *SafeMap[K, V]) ForEach(fn func(K, V) bool) {
	for shard := range m.maps {
		m.locks[shard].RLock()
		for k, v := range m.maps[shard] {
			if !fn(k, v) {
				m.locks[shard].RUnlock()
				return
			}
		}
		m.locks[shard].RUnlock()
	}
}

// ToMap 将 SafeMap 转换为一个简单的 map
func (m *SafeMap[K, V]) ToMap() map[K]V {
	// 创建一个新的简单 map
	simpleMap := make(map[K]V)

	// 遍历所有分片
	for shard := range m.maps {
		m.locks[shard].RLock() // 加读锁，确保线程安全
		for k, v := range m.maps[shard] {
			simpleMap[k] = v // 合并到简单 map 中
		}
		m.locks[shard].RUnlock() // 释放读锁
	}

	return simpleMap
}

// getShard 根据键获取对应的分片
func (m *SafeMap[K, V]) getShard(key K) uint64 {
	hash := fnv32(fmt.Sprintf("%v", key))
	return uint64(hash) % m.shardCount
}

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	keyLength := len(key)
	for i := 0; i < keyLength; i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}
