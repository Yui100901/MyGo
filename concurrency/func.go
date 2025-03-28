package concurrency

import (
	"fmt"
	"sync"
)

//
// @Author yfy2001
// @Date 2025/3/27 21 45
//

// ConcurrentFunc 定义一个接口
type ConcurrentFunc[T any] interface {
	ToConcurrentFunc() func() (T, error)
}

type TaskResult[T any] struct {
	Index int   // 函数索引
	Value T     // 成功时的返回值
	Err   error // 失败时的错误（包括 panic）
}

// ConcurrentRun 并发执行传入的函数集合，返回结果通道和错误通道。
// 每个函数的结果会发送到结果通道，错误（包括 panic 转换的错误）发送到错误通道。
// 当所有函数执行完成后，通道会自动关闭。
func ConcurrentRun[T any](funcs []func() (T, error)) <-chan TaskResult[T] {
	results := make(chan TaskResult[T], len(funcs))
	var wg sync.WaitGroup
	wg.Add(len(funcs))

	for i, fn := range funcs {
		go func(index int, f func() (T, error)) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					// 将 panic 转为 error
					results <- TaskResult[T]{
						Index: index,
						Err:   fmt.Errorf("panic: %v", r),
					}
				}
			}()

			// 执行函数并发送结果
			res, err := f()
			results <- TaskResult[T]{
				Index: index,
				Value: res,
				Err:   err,
			}
		}(i, fn)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}
