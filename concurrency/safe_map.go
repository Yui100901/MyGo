package concurrency

import (
	"fmt"
	"hash/maphash"
	"sync"
)

//
// @Author yfy2001
// @Date 2024/12/20 09 38
//

const defaultShardCount = 32

var seed = maphash.MakeSeed()

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
	safeMap.SetBatch(m)
	return safeMap
}

// Set 设置键值对
func (m *SafeMap[K, V]) Set(key K, value V) {
	shard := m.getShard(key)
	m.locks[shard].Lock()
	m.maps[shard][key] = value
	m.locks[shard].Unlock()
}

// SetBatch 批量设置键值对
func (m *SafeMap[K, V]) SetBatch(pairs map[K]V) {
	// 按分片分组数据
	shardData := make([]map[K]V, m.shardCount)
	for i := range shardData {
		shardData[i] = make(map[K]V)
	}

	// 将数据分配到对应的分片
	for k, v := range pairs {
		shard := m.getShard(k)
		shardData[shard][k] = v
	}

	// 并发处理每个分片
	var wg sync.WaitGroup
	for shard := uint64(0); shard < m.shardCount; shard++ {
		if len(shardData[shard]) == 0 {
			continue
		}

		wg.Add(1)
		go func(shardIndex uint64, data map[K]V) {
			defer wg.Done()
			m.locks[shardIndex].Lock()
			for k, v := range data {
				m.maps[shardIndex][k] = v
			}
			m.locks[shardIndex].Unlock()
		}(shard, shardData[shard])
	}
	wg.Wait()
}

// SetBatchSlice 批量设置键值对（切片形式）
func (m *SafeMap[K, V]) SetBatchSlice(keys []K, values []V) error {
	if len(keys) != len(values) {
		return fmt.Errorf("keys and values length mismatch: %d vs %d", len(keys), len(values))
	}

	// 按分片分组数据
	shardData := make([][]struct {
		key   K
		value V
	}, m.shardCount)

	// 将数据分配到对应的分片
	for i, key := range keys {
		shard := m.getShard(key)
		shardData[shard] = append(shardData[shard], struct {
			key   K
			value V
		}{key, values[i]})
	}

	// 并发处理每个分片
	var wg sync.WaitGroup
	for shard := uint64(0); shard < m.shardCount; shard++ {
		if len(shardData[shard]) == 0 {
			continue
		}

		wg.Add(1)
		go func(shardIndex uint64, data []struct {
			key   K
			value V
		}) {
			defer wg.Done()
			m.locks[shardIndex].Lock()
			for _, item := range data {
				m.maps[shardIndex][item.key] = item.value
			}
			m.locks[shardIndex].Unlock()
		}(shard, shardData[shard])
	}
	wg.Wait()
	return nil
}

// Get 获取键对应的值，不存在时返回nil
func (m *SafeMap[K, V]) Get(key K) (V, bool) {
	shard := m.getShard(key)
	m.locks[shard].RLock()
	value, ok := m.maps[shard][key]
	m.locks[shard].RUnlock()
	return value, ok
}

// GetBatch 批量获取键对应的值
func (m *SafeMap[K, V]) GetBatch(keys []K) map[K]V {
	result := make(map[K]V, len(keys))

	// 按分片分组键
	shardKeys := make([][]K, m.shardCount)
	for _, key := range keys {
		shard := m.getShard(key)
		shardKeys[shard] = append(shardKeys[shard], key)
	}

	// 并发处理每个分片
	var mu sync.Mutex
	var wg sync.WaitGroup

	for shard := uint64(0); shard < m.shardCount; shard++ {
		if len(shardKeys[shard]) == 0 {
			continue
		}

		wg.Add(1)
		go func(shardIndex uint64, keys []K) {
			defer wg.Done()
			m.locks[shardIndex].RLock()

			shardResult := make(map[K]V)
			for _, key := range keys {
				if value, ok := m.maps[shardIndex][key]; ok {
					shardResult[key] = value
				}
			}
			m.locks[shardIndex].RUnlock()

			// 合并结果需要加锁
			mu.Lock()
			for k, v := range shardResult {
				result[k] = v
			}
			mu.Unlock()
		}(shard, shardKeys[shard])
	}
	wg.Wait()

	return result
}

func (m *SafeMap[K, V]) MustGet(key K) V {
	shard := m.getShard(key)
	m.locks[shard].RLock()
	value, ok := m.maps[shard][key]
	m.locks[shard].RUnlock()
	if !ok {
		panic(fmt.Sprintf("SafeMap MustGet: key %v not exists", key))
	}
	return value
}

// GetOr 获取键对应的值，不存在时返回给定的默认值
func (m *SafeMap[K, V]) GetOr(key K, dV V) V {
	shard := m.getShard(key)
	m.locks[shard].RLock()
	value, ok := m.maps[shard][key]
	m.locks[shard].RUnlock()
	if ok {
		return value
	}
	return dV
}

// GetOrElse 获取键对应的值，不存在时调用给定的函数并返回其结果
func (m *SafeMap[K, V]) GetOrElse(key K, fn func() V) V {
	shard := m.getShard(key)
	m.locks[shard].RLock()
	value, ok := m.maps[shard][key]
	m.locks[shard].RUnlock()
	if ok {
		return value
	}
	return fn()
}

// Update 更新某个键的值，当keep返回为false时删除该键
func (m *SafeMap[K, V]) Update(key K, updater func(old V) (V, bool)) {
	shard := m.getShard(key)
	m.locks[shard].Lock()
	old, _ := m.maps[shard][key]
	newVal, keep := updater(old)
	if !keep {
		delete(m.maps[shard], key)
	} else {
		m.maps[shard][key] = newVal
	}
	m.locks[shard].Unlock()
}

// UpdateBatch 批量更新键值对
func (m *SafeMap[K, V]) UpdateBatch(updates map[K]func(V) (V, bool)) {
	// 按分片分组更新操作
	shardUpdates := make([]map[K]func(V) (V, bool), m.shardCount)
	for i := range shardUpdates {
		shardUpdates[i] = make(map[K]func(V) (V, bool))
	}

	// 将更新操作分配到对应的分片
	for k, updateFn := range updates {
		shard := m.getShard(k)
		shardUpdates[shard][k] = updateFn
	}

	// 并发处理每个分片
	var wg sync.WaitGroup
	for shard := uint64(0); shard < m.shardCount; shard++ {
		if len(shardUpdates[shard]) == 0 {
			continue
		}

		wg.Add(1)
		go func(shardIndex uint64, updates map[K]func(V) (V, bool)) {
			defer wg.Done()
			m.locks[shardIndex].Lock()
			for key, updateFn := range updates {
				old, _ := m.maps[shardIndex][key]
				newVal, keep := updateFn(old)
				if !keep {
					delete(m.maps[shardIndex], key)
				} else {
					m.maps[shardIndex][key] = newVal
				}
			}
			m.locks[shardIndex].Unlock()
		}(shard, shardUpdates[shard])
	}
	wg.Wait()
}

// Pop 返回并删除某个键
func (m *SafeMap[K, V]) Pop(key K) (V, bool) {
	shard := m.getShard(key)
	m.locks[shard].Lock()
	value, ok := m.maps[shard][key]
	if ok {
		delete(m.maps[shard], key)
	}
	m.locks[shard].Unlock()
	return value, ok
}

// Delete 删除键对应的值
func (m *SafeMap[K, V]) Delete(key K) {
	shard := m.getShard(key)
	m.locks[shard].Lock()
	delete(m.maps[shard], key)
	m.locks[shard].Unlock()
}

// DeleteBatch 批量删除键
func (m *SafeMap[K, V]) DeleteBatch(keys []K) {
	// 按分片分组键
	shardKeys := make([][]K, m.shardCount)
	for _, key := range keys {
		shard := m.getShard(key)
		shardKeys[shard] = append(shardKeys[shard], key)
	}

	// 并发处理每个分片
	var wg sync.WaitGroup
	for shard := uint64(0); shard < m.shardCount; shard++ {
		if len(shardKeys[shard]) == 0 {
			continue
		}

		wg.Add(1)
		go func(shardIndex uint64, keys []K) {
			defer wg.Done()
			m.locks[shardIndex].Lock()
			for _, key := range keys {
				delete(m.maps[shardIndex], key)
			}
			m.locks[shardIndex].Unlock()
		}(shard, shardKeys[shard])
	}
	wg.Wait()
}

// DeleteIf 根据传入的判断函数删除满足条件的键值对
func (m *SafeMap[K, V]) DeleteIf(predicate func(K, V) bool) {
	var wg sync.WaitGroup

	for shard := range m.maps {
		wg.Add(1)
		go func(shardIndex int) {
			defer wg.Done()
			m.locks[shardIndex].Lock()
			for k, v := range m.maps[shardIndex] {
				if predicate(k, v) {
					delete(m.maps[shardIndex], k)
				}
			}
			m.locks[shardIndex].Unlock()
		}(shard)
	}

	wg.Wait()
}

// Has 判断某个键是否存在
func (m *SafeMap[K, V]) Has(key K) bool {
	shard := m.getShard(key)
	m.locks[shard].RLock()
	_, ok := m.maps[shard][key]
	m.locks[shard].RUnlock()
	return ok
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

// ForEachAsync 并发遍历
func (m *SafeMap[K, V]) ForEachAsync(fn func(K, V)) {
	var wg sync.WaitGroup
	for shard := range m.maps {
		wg.Add(1)
		go func(shard int) {
			defer wg.Done()
			m.locks[shard].RLock()
			for k, v := range m.maps[shard] {
				fn(k, v)
			}
			m.locks[shard].RUnlock()
		}(shard)
	}
	wg.Wait()
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
	hash := hashKey(fmt.Sprintf("%v", key))
	return hash % m.shardCount
}

func hashKey[K comparable](key K) uint64 {
	var h maphash.Hash
	h.SetSeed(seed)
	switch v := any(key).(type) {
	case string:
		h.WriteString(v)
	case []byte:
		h.Write(v)
	default:
		fmt.Fprintf(&h, "%v", v) // fallback
	}
	return h.Sum64()
}
