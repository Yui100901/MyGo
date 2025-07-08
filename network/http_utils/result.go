package http_utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

//
// @Author yfy2001
// @Date 2025/7/8 10 22
//

// HTTPResult 封装 HTTP 响应
type HTTPResult struct {
	response *http.Response
	err      error
}

func ReadBodyByte(response *http.Response) ([]byte, error) {
	if response.Body != nil {
		defer response.Body.Close()
	} else {
		return nil, fmt.Errorf("nil response body")
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	return data, nil
}

func (r *HTTPResult) Response() (*http.Response, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.response, nil
}

// ParseJSON 解析 JSON 响应
func (r *HTTPResult) ParseJSON(target interface{}) error {
	if r.err != nil {
		return r.err
	}
	data, err := ReadBodyByte(r.response)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	return json.Unmarshal(data, target)
}

// GetBodyString 获取响应体字符串
func (r *HTTPResult) GetBodyString() (string, error) {
	data, err := ReadBodyByte(r.response)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}
	return string(data), nil
}

func (r *HTTPResult) SaveToFile(path string) error {
	if r.err != nil {
		return r.err
	}
	defer r.response.Body.Close()

	// 检查状态码
	if r.response.StatusCode < 200 || r.response.StatusCode >= 300 {
		return fmt.Errorf("非成功状态码: %d %s", r.response.StatusCode, r.response.Status)
	}

	// 创建目标文件
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 写入数据
	size, err := io.Copy(file, r.response.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	Logger.Printf("文件保存成功: %s | 大小: %d bytes", path, size)
	return nil
}
