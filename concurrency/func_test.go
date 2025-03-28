package concurrency

import (
	"errors"
	"strings"
	"testing"
	"time"
)

//
// @Author yfy2001
// @Date 2025/3/28 13 13
//

func TestConcurrentRun_Success(t *testing.T) {
	funcs := []func() (int, error){
		func() (int, error) { return 1, nil },
		func() (int, error) { return 2, nil },
		func() (int, error) { return 3, nil },
	}

	results := ConcurrentRun(funcs)
	resultMap := make(map[int]TaskResult[int])

	for res := range results {
		resultMap[res.Index] = res
	}

	if len(resultMap) != len(funcs) {
		t.Fatalf("Expected %d results, got %d", len(funcs), len(resultMap))
	}

	for i := 0; i < len(funcs); i++ {
		res, exists := resultMap[i]
		if !exists {
			t.Errorf("Result for index %d missing", i)
			continue
		}
		expected, _ := funcs[i]()
		if res.Value != expected || res.Err != nil {
			t.Errorf("Index %d: expected (%d, nil), got (%d, %v)",
				i, expected, res.Value, res.Err)
		}
	}
}

func TestConcurrentRun_ErrorHandling(t *testing.T) {
	testErr := errors.New("intentional error")
	funcs := []func() (string, error){
		func() (string, error) { return "success", nil },
		func() (string, error) { return "", testErr },
		func() (string, error) { return "another", nil },
	}

	results := ConcurrentRun(funcs)
	errorCount := 0

	for res := range results {
		if res.Index == 1 {
			if res.Err != testErr {
				t.Errorf("Expected error %v, got %v", testErr, res.Err)
			}
			errorCount++
		} else if res.Err != nil {
			t.Errorf("Unexpected error at index %d: %v", res.Index, res.Err)
		}
	}

	if errorCount != 1 {
		t.Errorf("Expected 1 error, got %d", errorCount)
	}
}

func TestConcurrentRun_PanicHandling(t *testing.T) {
	funcs := []func() (bool, error){
		func() (bool, error) { return true, nil },
		func() (bool, error) { panic("test panic") },
		func() (bool, error) { return false, nil },
	}

	results := ConcurrentRun(funcs)
	panicFound := false

	for res := range results {
		if res.Index == 1 {
			if res.Err == nil || !strings.Contains(res.Err.Error(), "panic: test panic") {
				t.Errorf("Expected panic error, got %v", res.Err)
			}
			panicFound = true
		} else if res.Err != nil {
			t.Errorf("Unexpected error at index %d: %v", res.Index, res.Err)
		}
	}

	if !panicFound {
		t.Error("Panic result not found")
	}
}

func TestConcurrentRun_Concurrency(t *testing.T) {
	const (
		taskDuration = 100 * time.Millisecond
		numTasks     = 5
	)

	// 创建模拟耗时任务
	funcs := make([]func() (int, error), numTasks)
	for i := 0; i < numTasks; i++ {
		funcs[i] = func() (int, error) {
			time.Sleep(taskDuration)
			return 0, nil
		}
	}

	// 测试并发版本
	start := time.Now()
	results := ConcurrentRun(funcs)
	for range results { // 等待所有结果完成
	}
	concurrentTime := time.Since(start)

	// 测试顺序执行
	start = time.Now()
	for _, fn := range funcs {
		fn()
	}
	sequentialTime := time.Since(start)

	t.Logf("Concurrent time: %v, Sequential time: %v", concurrentTime, sequentialTime)

	// 验证并发执行时间显著短于顺序执行
	maxAllowed := taskDuration + 50*time.Millisecond
	if concurrentTime > maxAllowed {
		t.Errorf("Concurrent execution took too long: %v (max allowed %v)",
			concurrentTime, maxAllowed)
	}

	// 验证顺序执行时间符合预期
	minExpected := time.Duration(numTasks) * taskDuration
	if sequentialTime < minExpected {
		t.Errorf("Sequential execution too fast: %v (minimum expected %v)",
			sequentialTime, minExpected)
	}
}

func TestConcurrentRun_ChannelClosing(t *testing.T) {
	funcs := []func() (int, error){
		func() (int, error) { return 1, nil },
	}

	results := ConcurrentRun(funcs)
	_, ok := <-results // 第一次读取应该成功
	if !ok {
		t.Fatal("Channel closed too early")
	}

	_, ok = <-results // 第二次读取应该失败
	if ok {
		t.Error("Channel not closed after all results")
	}
}
