package collection

//
// @Author yfy2001
// @Date 2025/4/7 13 55
//

func MapFilter[K comparable, V any](data map[K]V, predicate func(K, V) bool) map[K]V {
	var result map[K]V
	for k, v := range data {
		if predicate(k, v) {
			result[k] = v
		}
	}
	return result
}
