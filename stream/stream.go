package stream

import (
	"iter"
	"slices"
	"sort"
	"sync"
)

//
// @Author yfy2001
// @Date 2025/4/9 16 50
//

type Stream[E any] struct {
	it iter.Seq[E]
}

// NewStream 创建一个新的流
func NewStream[E any](it iter.Seq[E]) *Stream[E] {
	return &Stream[E]{it: it}
}

// FromSlice 从切片创建流
func FromSlice[E any](slice []E) *Stream[E] {
	return &Stream[E]{
		it: slices.Values(slice),
	}
}

// FromMap 从映射创建流
func FromMap[K comparable, V any](m map[K]V) *Stream[MapEntry[K, V]] {
	entries := make([]MapEntry[K, V], 0, len(m))
	for k, v := range m {
		entries = append(entries, MapEntry[K, V]{Key: k, Value: v})
	}
	return &Stream[MapEntry[K, V]]{
		it: slices.Values(entries),
	}
}

type MapEntry[K comparable, V any] struct {
	Key   K
	Value V
}

//中间函数，返回流

// Map 对流中的每个元素应用映射函数，返回新的流
func Map[E any, R any](s *Stream[E], mapper func(E) R) *Stream[R] {
	return NewStream(func(yield func(R) bool) {
		s.it(func(elem E) bool {
			return yield(mapper(elem))
		})
	})
}

// Concat 连接多个流
func Concat[E any](streams ...*Stream[E]) *Stream[E] {
	return NewStream(func(yield func(E) bool) {
		for _, stream := range streams {
			// 处理当前流中的元素
			continueProcessing := true
			stream.it(func(elem E) bool {
				if !yield(elem) {
					continueProcessing = false
					return false // 停止当前流的处理
				}
				return true
			})
			if !continueProcessing {
				return // 停止处理后续流
			}
		}
	})
}

// Filter 筛选满足条件的元素
func (s *Stream[E]) Filter(predicate func(E) bool) *Stream[E] {
	return NewStream(func(yield func(E) bool) {
		s.it(func(elem E) bool {
			if predicate(elem) {
				return yield(elem)
			}
			return true
		})
	})
}

// Sorted 按照比较函数对流中的元素排序
func (s *Stream[E]) Sorted(less func(E, E) bool) *Stream[E] {
	return NewStream(func(yield func(E) bool) {
		var elements []E
		s.it(func(elem E) bool {
			elements = append(elements, elem)
			return true
		})
		sort.Slice(elements, func(i, j int) bool {
			return less(elements[i], elements[j])
		})
		for _, elem := range elements {
			if !yield(elem) {
				return
			}
		}
	})
}

// Distinct 去除流中的重复元素
func (s *Stream[E]) Distinct() *Stream[E] {
	return NewStream(func(yield func(E) bool) {
		seen := sync.Map{}
		s.it(func(elem E) bool {
			if _, exists := seen.LoadOrStore(elem, true); !exists {
				return yield(elem)
			}
			return true
		})
	})
}

// Peek 对每个元素执行动作，但不改变流内容
func (s *Stream[E]) Peek(action func(E)) *Stream[E] {
	return NewStream(func(yield func(E) bool) {
		s.it(func(elem E) bool {
			action(elem)
			return yield(elem)
		})
	})
}

// Limit 限制流中元素的数量
func (s *Stream[E]) Limit(count int) *Stream[E] {
	return NewStream(func(yield func(E) bool) {
		i := 0
		s.it(func(elem E) bool {
			if i < count {
				i++
				return yield(elem)
			}
			return false
		})
	})
}

// Skip 跳过流中指定数量的元素
func (s *Stream[E]) Skip(n int) *Stream[E] {
	return NewStream(func(yield func(E) bool) {
		i := 0
		s.it(func(elem E) bool {
			if i >= n { // 当跳过的元素达到指定数量时，开始处理后续元素
				return yield(elem)
			}
			i++
			return true // 跳过当前元素
		})
	})
}

//终止函数，返回某个特定集合或某个符合条件的元素

// FindFirst 查找第一个满足条件的元素
func (s *Stream[E]) FindFirst(predicate func(E) bool) *E {
	var first *E
	s.it(func(elem E) bool {
		if predicate(elem) {
			first = &elem
			return false // 找到第一个符合条件的元素后停止
		}
		return true
	})
	return first
}

// Min 返回流中元素的最小值
func (s *Stream[E]) Min(less func(E, E) bool) *E {
	var minimum *E
	s.it(func(elem E) bool {
		if minimum == nil || less(elem, *minimum) {
			minimum = &elem
		}
		return true
	})
	return minimum
}

// Max 返回流中元素的最大值
func (s *Stream[E]) Max(less func(E, E) bool) *E {
	var maximum *E
	s.it(func(elem E) bool {
		if maximum == nil || less(*maximum, elem) {
			maximum = &elem
		}
		return true
	})
	return maximum
}

// Collect 通用的收集函数，支持用户自定义逻辑
func Collect[E any, R any](s *Stream[E], collector func(iter.Seq[E]) R) R {
	return collector(s.it)
}

// ToSlice 收集元素为切片
func (s *Stream[E]) ToSlice() []E {
	result := make([]E, 0)
	s.it(func(elem E) bool {
		result = append(result, elem)
		return true
	})
	return result
}

// ToMap 将流中的元素转换为 map
func ToMap[E any, K comparable, V any](s *Stream[E], mapper func(E) (K, V)) map[K]V {
	result := make(map[K]V)
	s.it(func(elem E) bool {
		key, value := mapper(elem)
		result[key] = value
		return true // 继续处理流中的下一个元素
	})
	return result
}

// GroupBy 根据分类器将元素分组
func GroupBy[E any, K comparable](s *Stream[E], classifier func(E) K) map[K][]E {
	result := make(map[K][]E)
	s.it(func(elem E) bool {
		key := classifier(elem)
		result[key] = append(result[key], elem)
		return true
	})
	return result
}
