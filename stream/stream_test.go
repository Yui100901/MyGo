package stream

import (
	"fmt"
	"reflect"
	"testing"
)

//
// @Author yfy2001
// @Date 2025/4/10 10 16
//

// 定义 User 结构体
type User struct {
	ID   int
	Name string
	Age  int
}

func TestStreamFunctions(t *testing.T) {
	data := []int{1, 2, 3, 4, 5}
	streamData := FromSlice(data)

	t.Run("TestFromSliceAndToSlice", func(t *testing.T) {
		if !reflect.DeepEqual(data, streamData.ToSlice()) {
			t.Errorf("Expected %v, but got %v", data, streamData.ToSlice())
		}
	})

	t.Run("TestMapFunction", func(t *testing.T) {
		mappedStream := Map(streamData, func(x int) string {
			return fmt.Sprintf("Number: %d", x)
		})
		mappedResult := mappedStream.ToSlice()
		expectedMapped := []string{"Number: 1", "Number: 2", "Number: 3", "Number: 4", "Number: 5"}
		if !reflect.DeepEqual(mappedResult, expectedMapped) {
			t.Errorf("Expected %v, but got %v", expectedMapped, mappedResult)
		}
	})

	t.Run("TestFilterFunction", func(t *testing.T) {
		filteredStream := streamData.Filter(func(x int) bool {
			return x%2 == 0
		})
		filteredResult := filteredStream.ToSlice()
		expectedFiltered := []int{2, 4}
		if !reflect.DeepEqual(filteredResult, expectedFiltered) {
			t.Errorf("Expected %v, but got %v", expectedFiltered, filteredResult)
		}
	})

	t.Run("TestConcatFunction", func(t *testing.T) {
		stream1 := FromSlice([]int{1, 2, 3})
		stream2 := FromSlice([]int{4, 5, 6})
		concatenatedStream := Concat(stream1, stream2)
		concatenatedResult := concatenatedStream.ToSlice()
		expectedConcatenated := []int{1, 2, 3, 4, 5, 6}
		if !reflect.DeepEqual(concatenatedResult, expectedConcatenated) {
			t.Errorf("Expected %v, but got %v", expectedConcatenated, concatenatedResult)
		}
	})

	t.Run("TestSortedFunction", func(t *testing.T) {
		unsortedStream := FromSlice([]int{5, 1, 4, 2, 3})
		sortedStream := unsortedStream.Sorted(func(a, b int) bool {
			return a < b
		})
		sortedResult := sortedStream.ToSlice()
		expectedSorted := []int{1, 2, 3, 4, 5}
		if !reflect.DeepEqual(sortedResult, expectedSorted) {
			t.Errorf("Expected %v, but got %v", expectedSorted, sortedResult)
		}
	})

	t.Run("TestDistinctFunction", func(t *testing.T) {
		duplicatesStream := FromSlice([]int{1, 2, 2, 3, 3, 3, 4})
		distinctStream := duplicatesStream.Distinct()
		distinctResult := distinctStream.ToSlice()
		expectedDistinct := []int{1, 2, 3, 4}
		if !reflect.DeepEqual(distinctResult, expectedDistinct) {
			t.Errorf("Expected %v, but got %v", expectedDistinct, distinctResult)
		}
	})

	t.Run("TestFindFirstFunction", func(t *testing.T) {
		firstResult := streamData.FindFirst(func(x int) bool {
			return x > 3
		})
		if firstResult == nil || *firstResult != 4 {
			t.Errorf("Expected %v, but got %v", 4, firstResult)
		}
	})

	t.Run("TestLimitAndSkipFunctions", func(t *testing.T) {
		limitedStream := streamData.Limit(3)
		limitedResult := limitedStream.ToSlice()
		expectedLimited := []int{1, 2, 3}
		if !reflect.DeepEqual(limitedResult, expectedLimited) {
			t.Errorf("Expected %v, but got %v", expectedLimited, limitedResult)
		}

		skippedStream := streamData.Skip(3)
		skippedResult := skippedStream.ToSlice()
		expectedSkipped := []int{4, 5}
		if !reflect.DeepEqual(skippedResult, expectedSkipped) {
			t.Errorf("Expected %v, but got %v", expectedSkipped, skippedResult)
		}
	})

	t.Run("TestMinAndMaxFunctions", func(t *testing.T) {
		minResult := streamData.Min(func(a, b int) bool {
			return a < b
		})
		if minResult == nil || *minResult != 1 {
			t.Errorf("Expected %v, but got %v", 1, minResult)
		}

		maxResult := streamData.Max(func(a, b int) bool {
			return a < b
		})
		if maxResult == nil || *maxResult != 5 {
			t.Errorf("Expected %v, but got %v", 5, maxResult)
		}
	})

	t.Run("TestGroupByFunction", func(t *testing.T) {
		users := []User{
			{ID: 1, Name: "Alice", Age: 25},
			{ID: 2, Name: "Bob", Age: 30},
			{ID: 3, Name: "Charlie", Age: 25},
		}
		groupedResult := GroupBy(FromSlice(users), func(u User) int {
			return u.Age
		})
		if len(groupedResult[25]) != 2 || len(groupedResult[30]) != 1 {
			t.Errorf("Expected groups {25:2, 30:1}, but got %v", groupedResult)
		}
	})

	t.Run("TestToMapFunction", func(t *testing.T) {
		users := []User{
			{ID: 1, Name: "Alice", Age: 25},
			{ID: 2, Name: "Bob", Age: 30},
			{ID: 3, Name: "Charlie", Age: 25},
		}
		userMapResult := ToMap(FromSlice(users), func(u User) (int, string) {
			return u.ID, u.Name
		})
		expectedMap := map[int]string{
			1: "Alice",
			2: "Bob",
			3: "Charlie",
		}
		if !reflect.DeepEqual(userMapResult, expectedMap) {
			t.Errorf("Expected %v, but got %v", expectedMap, userMapResult)
		}
	})
}
