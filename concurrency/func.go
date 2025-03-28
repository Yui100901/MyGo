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

// ConcurrentRun 异步执行传入的函数集合，返回结果通道和错误通道。
// 每个函数的结果会发送到结果通道，错误（包括 panic 转换的错误）发送到错误通道。
// 当所有函数执行完成后，通道会自动关闭。
func ConcurrentRun[T any](funcs ...func() (T, error)) (<-chan T, <-chan error) {
	results := make(chan T, len(funcs))
	errors := make(chan error, len(funcs))

	var wg sync.WaitGroup
	wg.Add(len(funcs))

	processFunc := func(f func() (T, error)) {
		defer wg.Done()

		// 捕获 panic 并添加堆栈跟踪
		defer func() {
			if r := recover(); r != nil {
				errMsg := fmt.Sprintf("panic: %v", r)
				errors <- fmt.Errorf(errMsg)
			}
		}()

		// 执行函数并处理结果
		if res, err := f(); err != nil {
			errors <- err
		} else {
			results <- res
		}
	}

	for _, fn := range funcs {
		go processFunc(fn)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	return results, errors
}
