package db

//
// @Author yfy2001
// @Date 2025/8/18 10 16
//

// Result 查询结果统一封装
type Result[T any] struct {
	Data    T     // 数据（单条或列表）
	Rows    int64 // 影响行数（写操作）
	Err     error // 错误信息
	Success bool  // 是否成功
}

func Ok[T any](data T, rows int64) *Result[T] {
	return &Result[T]{
		Data:    data,
		Rows:    rows,
		Success: true,
	}
}

// Fail 创建失败结果
func Fail[T any](err error) *Result[T] {
	return &Result[T]{
		Err:     err,
		Success: false,
	}
}
