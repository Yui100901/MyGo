package concurrency

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

//
// @Author yfy2001
// @Date 2025/3/28 13 13
//

func TestConcurrentRun(t *testing.T) {
	tests := []struct {
		name      string
		funcs     []func() (string, error)
		wantRes   []string
		wantErrs  []string
		sleepTime time.Duration
	}{
		{
			name: "normal execution",
			funcs: []func() (string, error){
				func() (string, error) { time.Sleep(1 * time.Second); return "Task 1", nil },
				func() (string, error) { time.Sleep(500 * time.Millisecond); return "Task 2", nil },
			},
			wantRes:   []string{"Task 1", "Task 2"},
			wantErrs:  []string{},
			sleepTime: 2 * time.Second,
		},
		{
			name: "error handling",
			funcs: []func() (string, error){
				func() (string, error) { time.Sleep(1 * time.Second); return "", fmt.Errorf("error 1") },
				func() (string, error) { return "", fmt.Errorf("error 2") },
			},
			wantRes:   []string{},
			wantErrs:  []string{"error 1", "error 2"},
			sleepTime: 2 * time.Second,
		},
		{
			name: "panic handling",
			funcs: []func() (string, error){
				func() (string, error) { panic("panic occurred") },
			},
			wantRes:   []string{},
			wantErrs:  []string{"panic: panic occurred"},
			sleepTime: 1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, errors := ConcurrentRun(tt.funcs...)

			// 使用通道接收结果
			var res []string
			go func() {
				for r := range results {
					res = append(res, r)
				}
			}()

			// 使用通道接收错误
			var errs []string
			go func() {
				for e := range errors {
					errs = append(errs, e.Error())
				}
			}()

			// 等待测试完成
			time.Sleep(tt.sleepTime)

			// 验证结果
			if len(res) != len(tt.wantRes) {
				t.Errorf("expected results %v, got %v", tt.wantRes, res)
			}
			for _, r := range tt.wantRes {
				found := false
				for _, actual := range res {
					if r == actual {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected result %v not found in %v", r, res)
				}
			}

			// 验证错误
			if len(errs) != len(tt.wantErrs) {
				t.Errorf("expected errors %v, got %v", tt.wantErrs, errs)
			}
			for _, e := range tt.wantErrs {
				found := false
				for _, actual := range errs {
					if e == actual {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error %v not found in %v", e, errs)
				}
			}
		})
	}
}

func TestConcurrentRun2(t *testing.T) {
	// 定义要下载的网页 URL
	urls := []string{
		"https://www.example.com",
		"https://www.google.com",
		"https://www.bing.com",
		"https://www.invalid-url.com", // 模拟一个错误的 URL
	}

	// 为每个 URL 创建一个下载函数
	funcs := make([]func() (string, error), len(urls))
	for i, url := range urls {
		funcs[i] = func(u string) func() (string, error) {
			return func() (string, error) {
				resp, err := http.Get(u)
				if err != nil {
					return "", fmt.Errorf("failed to download %s: %v", u, err)
				}
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return "", fmt.Errorf("failed to read response from %s: %v", u, err)
				}

				return string(body), nil
			}
		}(url)
	}

	// 并发执行下载任务
	results, errors := ConcurrentRun(funcs...)

	// 处理下载结果
	go func() {
		for res := range results {
			fmt.Println("Page content received:", len(res), "bytes")
		}
	}()

	// 处理错误
	go func() {
		for err := range errors {
			fmt.Println("Error:", err)
		}
	}()

	// 防止主线程过早退出
	time.Sleep(5 * time.Second)
}
