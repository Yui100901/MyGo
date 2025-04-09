package collection

//
// @Author yfy2001
// @Date 2025/4/7 13 39
//

// SliceFilter 条件过滤
func SliceFilter[T any](data []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range data {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}
