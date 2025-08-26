package concurrency

import (
	"fmt"
	"sync"
)

//
// @Author yfy2001
// @Date 2025/8/26 15 12
//

// SafeBiMap 基于SafeMap实现的线程安全双向map
// 使用分片锁提供更好的并发性能
// 限制：双向map需要保证键值都唯一
// 占用较高且不适合频繁变更
type SafeBiMap[K comparable, V comparable] struct {
	forward  *SafeMap[K, V] // 正向映射 K -> V
	backward *SafeMap[V, K] // 反向映射 V -> K
	mu       sync.RWMutex   // 全局锁，用于保证双向映射的原子性
}

// NewSafeBiMapV2 创建一个新的基于SafeMap的双向map
func NewSafeBiMapV2[K comparable, V comparable](shardCount int) *SafeBiMap[K, V] {
	if shardCount <= 0 {
		shardCount = defaultShardCount
	}

	return &SafeBiMap[K, V]{
		forward:  NewSafeMap[K, V](shardCount),
		backward: NewSafeMap[V, K](shardCount),
	}
}

// NewSafeBiMapV2FromMap 从普通map创建双向map
func NewSafeBiMapV2FromMap[K comparable, V comparable](m map[K]V, shardCount int) (*SafeBiMap[K, V], error) {
	// 检查值是否有重复
	valueSet := make(map[V]bool)
	for _, v := range m {
		if valueSet[v] {
			return nil, fmt.Errorf("duplicate value found: %v", v)
		}
		valueSet[v] = true
	}

	biMap := NewSafeBiMapV2[K, V](shardCount)

	// 批量设置，利用SafeMap的批量操作
	forwardMap := make(map[K]V)
	backwardMap := make(map[V]K)

	for k, v := range m {
		forwardMap[k] = v
		backwardMap[v] = k
	}

	biMap.forward.SetBatch(forwardMap)
	biMap.backward.SetBatch(backwardMap)

	return biMap, nil
}

// Set 设置键值对，如果键或值已存在会覆盖原有映射
func (b *SafeBiMap[K, V]) Set(key K, value V) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 检查并清理可能存在的旧映射
	if oldValue, exists := b.forward.Get(key); exists {
		b.backward.Delete(oldValue)
	}

	if oldKey, exists := b.backward.Get(value); exists {
		b.forward.Delete(oldKey)
	}

	// 设置新映射
	b.forward.Set(key, value)
	b.backward.Set(value, key)
}

// TrySet 尝试设置键值对，如果键或值已存在则返回false
func (b *SafeBiMap[K, V]) TrySet(key K, value V) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 检查键或值是否已存在
	if b.forward.Has(key) || b.backward.Has(value) {
		return false
	}

	// 设置映射
	b.forward.Set(key, value)
	b.backward.Set(value, key)
	return true
}

// GetByKey 通过键获取值
func (b *SafeBiMap[K, V]) GetByKey(key K) (V, bool) {
	return b.forward.Get(key)
}

// GetByValue 通过值获取键
func (b *SafeBiMap[K, V]) GetByValue(value V) (K, bool) {
	return b.backward.Get(value)
}

// MustGetByKey 通过键获取值，不存在时panic
func (b *SafeBiMap[K, V]) MustGetByKey(key K) V {
	return b.forward.MustGet(key)
}

// MustGetByValue 通过值获取键，不存在时panic
func (b *SafeBiMap[K, V]) MustGetByValue(value V) K {
	return b.backward.MustGet(value)
}

// GetByKeyOr 通过键获取值，不存在时返回默认值
func (b *SafeBiMap[K, V]) GetByKeyOr(key K, defaultValue V) V {
	return b.forward.GetOr(key, defaultValue)
}

// GetByValueOr 通过值获取键，不存在时返回默认键
func (b *SafeBiMap[K, V]) GetByValueOr(value V, defaultKey K) K {
	return b.backward.GetOr(value, defaultKey)
}

// GetByKeyOrElse 通过键获取值，不存在时调用函数
func (b *SafeBiMap[K, V]) GetByKeyOrElse(key K, fn func() V) V {
	return b.forward.GetOrElse(key, fn)
}

// GetByValueOrElse 通过值获取键，不存在时调用函数
func (b *SafeBiMap[K, V]) GetByValueOrElse(value V, fn func() K) K {
	return b.backward.GetOrElse(value, fn)
}

// HasKey 检查键是否存在
func (b *SafeBiMap[K, V]) HasKey(key K) bool {
	return b.forward.Has(key)
}

// HasValue 检查值是否存在
func (b *SafeBiMap[K, V]) HasValue(value V) bool {
	return b.backward.Has(value)
}

// DeleteByKey 通过键删除映射
func (b *SafeBiMap[K, V]) DeleteByKey(key K) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	value, exists := b.forward.Get(key)
	if !exists {
		return false
	}

	b.forward.Delete(key)
	b.backward.Delete(value)
	return true
}

// DeleteByValue 通过值删除映射
func (b *SafeBiMap[K, V]) DeleteByValue(value V) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	key, exists := b.backward.Get(value)
	if !exists {
		return false
	}

	b.forward.Delete(key)
	b.backward.Delete(value)
	return true
}

// PopByKey 通过键弹出值（获取并删除）
func (b *SafeBiMap[K, V]) PopByKey(key K) (V, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	value, exists := b.forward.Pop(key)
	if !exists {
		return value, false
	}

	b.backward.Delete(value)
	return value, true
}

// PopByValue 通过值弹出键（获取并删除）
func (b *SafeBiMap[K, V]) PopByValue(value V) (K, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	key, exists := b.backward.Pop(value)
	if !exists {
		return key, false
	}

	b.forward.Delete(key)
	return key, true
}

// Update 更新某个键的值
func (b *SafeBiMap[K, V]) Update(key K, updater func(old V) (V, bool)) {
	b.mu.Lock()
	defer b.mu.Unlock()

	oldValue, exists := b.forward.Get(key)
	if exists {
		b.backward.Delete(oldValue)
	}

	b.forward.Update(key, func(old V) (V, bool) {
		newValue, keep := updater(old)
		if keep {
			// 检查新值是否与其他键冲突
			if conflictKey, valueExists := b.backward.Get(newValue); valueExists && !equals(conflictKey, key) {
				b.forward.Delete(conflictKey)
			}
			b.backward.Set(newValue, key)
		}
		return newValue, keep
	})
}

// Clear 清空所有映射
func (b *SafeBiMap[K, V]) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.forward.Clear()
	b.backward.Clear()
}

// Length 返回映射数量
func (b *SafeBiMap[K, V]) Length() int {
	return b.forward.Length()
}

// IsEmpty 检查是否为空
func (b *SafeBiMap[K, V]) IsEmpty() bool {
	return b.forward.Length() == 0
}

// Keys 返回所有键
func (b *SafeBiMap[K, V]) Keys() []K {
	return b.forward.Keys()
}

// Values 返回所有值
func (b *SafeBiMap[K, V]) Values() []V {
	return b.forward.Values()
}

// ForEach 遍历所有键值对
func (b *SafeBiMap[K, V]) ForEach(fn func(K, V) bool) {
	b.forward.ForEach(fn)
}

// ForEachAsync 并发遍历所有键值对
func (b *SafeBiMap[K, V]) ForEachAsync(fn func(K, V)) {
	b.forward.ForEachAsync(fn)
}

// ToMap 转换为普通的正向map
func (b *SafeBiMap[K, V]) ToMap() map[K]V {
	return b.forward.ToMap()
}

// ToReverseMap 转换为普通的反向map
func (b *SafeBiMap[K, V]) ToReverseMap() map[V]K {
	return b.backward.ToMap()
}

// Clone 创建一个副本
func (b *SafeBiMap[K, V]) Clone() *SafeBiMap[K, V] {
	b.mu.RLock()
	defer b.mu.RUnlock()

	forwardMap := b.forward.ToMap()
	clone, _ := NewSafeBiMapV2FromMap(forwardMap, int(b.forward.shardCount))
	return clone
}

// Merge 合并另一个双向map，冲突时以other为准
func (b *SafeBiMap[K, V]) Merge(other *SafeBiMap[K, V]) {
	b.mu.Lock()
	defer b.mu.Unlock()

	otherMap := other.ToMap()

	// 先清理冲突的映射
	for k, v := range otherMap {
		if oldValue, exists := b.forward.Get(k); exists {
			b.backward.Delete(oldValue)
		}
		if oldKey, exists := b.backward.Get(v); exists {
			b.forward.Delete(oldKey)
		}
	}

	// 批量设置新映射
	backwardMap := make(map[V]K)
	for k, v := range otherMap {
		backwardMap[v] = k
	}

	b.forward.SetBatch(otherMap)
	b.backward.SetBatch(backwardMap)
}

// 批量操作方法

// SetBatch 批量设置键值对
func (b *SafeBiMap[K, V]) SetBatch(pairs map[K]V) error {
	// 检查值是否有重复
	valueSet := make(map[V]bool)
	for _, v := range pairs {
		if valueSet[v] {
			return fmt.Errorf("duplicate value in batch: %v", v)
		}
		valueSet[v] = true
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// 清理可能冲突的旧映射
	for k, v := range pairs {
		if oldValue, exists := b.forward.Get(k); exists {
			b.backward.Delete(oldValue)
		}
		if oldKey, exists := b.backward.Get(v); exists {
			b.forward.Delete(oldKey)
		}
	}

	// 构建反向映射
	backwardPairs := make(map[V]K, len(pairs))
	for k, v := range pairs {
		backwardPairs[v] = k
	}

	// 批量设置
	b.forward.SetBatch(pairs)
	b.backward.SetBatch(backwardPairs)

	return nil
}

// GetBatchByKeys 批量通过键获取值
func (b *SafeBiMap[K, V]) GetBatchByKeys(keys []K) map[K]V {
	return b.forward.GetBatch(keys)
}

// GetBatchByValues 批量通过值获取键
func (b *SafeBiMap[K, V]) GetBatchByValues(values []V) map[V]K {
	return b.backward.GetBatch(values)
}

// DeleteBatchByKeys 批量通过键删除
func (b *SafeBiMap[K, V]) DeleteBatchByKeys(keys []K) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 先获取要删除的值
	valuesToDelete := make([]V, 0, len(keys))
	existingKeys := make([]K, 0, len(keys))

	for _, key := range keys {
		if value, exists := b.forward.Get(key); exists {
			valuesToDelete = append(valuesToDelete, value)
			existingKeys = append(existingKeys, key)
		}
	}

	// 批量删除
	b.forward.DeleteBatch(existingKeys)
	b.backward.DeleteBatch(valuesToDelete)

	return len(existingKeys)
}

// DeleteBatchByValues 批量通过值删除
func (b *SafeBiMap[K, V]) DeleteBatchByValues(values []V) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 先获取要删除的键
	keysToDelete := make([]K, 0, len(values))
	existingValues := make([]V, 0, len(values))

	for _, value := range values {
		if key, exists := b.backward.Get(value); exists {
			keysToDelete = append(keysToDelete, key)
			existingValues = append(existingValues, value)
		}
	}

	// 批量删除
	b.forward.DeleteBatch(keysToDelete)
	b.backward.DeleteBatch(existingValues)

	return len(existingValues)
}

// InvertedView 返回一个反向视图，交换键值的角色
func (b *SafeBiMap[K, V]) InvertedView() *SafeBiMap[V, K] {
	return &SafeBiMap[V, K]{
		forward:  b.backward,
		backward: b.forward,
		mu:       b.mu, // 共享同一个锁
	}
}

// String 返回字符串表示
func (b *SafeBiMap[K, V]) String() string {
	forwardMap := b.forward.ToMap()
	backwardMap := b.backward.ToMap()
	return fmt.Sprintf("SafeBiMap{forward: %v, backward: %v}", forwardMap, backwardMap)
}

// 辅助函数：比较两个值是否相等
func equals[T comparable](a, b T) bool {
	return a == b
}
