package random

import (
	"math/rand"
	"time"
)

//
// @Author yfy2001
// @Date 2025/1/21 10 25
//

func init() {
	rand.NewSource(time.Now().UnixNano())
}

// RandItemFromSlice 从给定的切片随机取一个值
func RandItemFromSlice[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	return slice[rand.Intn(len(slice))]
}

// RandSliceFromSlice 从给定的切片随机截取一段切片
func RandSliceFromSlice[T any](slice []T, num int, repeatable bool) []T {
	if num <= 0 || len(slice) == 0 {
		return slice
	}

	if !repeatable && num > len(slice) {
		num = len(slice)
	}

	result := make([]T, num)
	if repeatable {
		for i := range result {
			result[i] = slice[rand.Intn(len(slice))]
		}
	} else {
		shuffled := make([]T, len(slice))
		copy(shuffled, slice)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})
		result = shuffled[:num]
	}
	return result
}
