package stream

import (
	"fmt"
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

func TestStreamChaining(t *testing.T) {
	// 创建一个 map
	userMap := map[int]User{
		1: {ID: 1, Name: "Alice", Age: 25},
		2: {ID: 2, Name: "Bob", Age: 30},
		3: {ID: 3, Name: "Alice", Age: 25}, // 重复数据
		4: {ID: 4, Name: "Eve", Age: 35},
	}
	result := Map(FromMap(userMap), func(e MapEntry[int, User]) User {
		return e.Value
	}).Filter(func(user User) bool { // 过滤年龄大于 28 的用户
		return user.Age > 28
	}).
		Distinct().                     // 去重
		Sorted(func(u1, u2 User) bool { // 按年龄排序
			return u1.Age < u2.Age
		}).
		ToSlice() // 转换为切片

	// 打印结果
	fmt.Println(result)

	// 验证结果是否正确
	expected := []User{
		{ID: 2, Name: "Bob", Age: 30},
		{ID: 4, Name: "Eve", Age: 35},
	}

	if len(result) != len(expected) {
		t.Errorf("Expected %d users, but got %d", len(expected), len(result))
	}

	for i, user := range result {
		if user != expected[i] {
			t.Errorf("Expected user %+v, but got %+v", expected[i], user)
		}
	}
}
