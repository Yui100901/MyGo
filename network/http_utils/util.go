package http_utils

import (
	"log"
	"net/http"
	"os"
	"strings"
)

//
// @Author yfy2001
// @Date 2025/7/7 14 21
//

var Logger *log.Logger

func init() {
	Logger = log.New(os.Stdout, "[HTTP] ", log.LstdFlags)
}

// 辅助函数

// fileSize 获取文件大小
func fileSize(file *os.File) int64 {
	info, err := file.Stat()
	if err != nil {
		return 0
	}
	return info.Size()
}

// isSensitiveHeader 判断是否是敏感头信息
func isSensitiveHeader(key string) bool {
	key = strings.ToLower(key)
	sensitiveHeaders := []string{
		"authorization",
		"cookie",
		"set-cookie",
		"proxy-authorization",
		"x-api-key",
		"x-access-token",
	}

	for _, header := range sensitiveHeaders {
		if key == header {
			return true
		}
	}
	return false
}

// logRequestHeaders 记录请求头 (过滤敏感信息)
func (c *HTTPClient) logRequestHeaders(req *http.Request) {
	Logger.Printf("请求头:")
	for key, values := range req.Header {
		// 过滤敏感头信息
		if isSensitiveHeader(key) {
			Logger.Printf("  %s: [过滤]", key)
		} else {
			Logger.Printf("  %s: %s", key, strings.Join(values, ", "))
		}
	}
}

// logResponseHeaders 记录响应头 (过滤敏感信息)
func (c *HTTPClient) logResponseHeaders(headers map[string]string) {
	Logger.Printf("响应头:")
	for key, value := range headers {
		// 过滤敏感头信息
		if isSensitiveHeader(key) {
			Logger.Printf("  %s: [过滤]", key)
		} else {
			Logger.Printf("  %s: %s", key, value)
		}
	}
}
