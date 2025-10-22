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

func ReadBodyBytes(response *http.Response) ([]byte, error) {
	if response != nil {
		if response.Body != nil {
			defer response.Body.Close()
		}
	} else {
		return nil, fmt.Errorf("nil response body")
	}
	bytesData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	return bytesData, nil
}

// ParseJSON 解析 JSON 响应
func ParseJSON(response *http.Response, target any) error {
	data, err := ReadBodyBytes(response)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	return json.Unmarshal(data, target)
}

// GetBodyString 获取响应体字符串
func GetBodyString(response *http.Response) (string, error) {
	data, err := ReadBodyBytes(response)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}
	return string(data), nil
}

func SaveToFile(response *http.Response, path string) error {
	defer response.Body.Close()

	// 检查状态码
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("非成功状态码: %d %s", response.StatusCode, response.Status)
	}

	// 创建目标文件
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 写入数据
	size, err := io.Copy(file, response.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	Logger.Printf("文件保存成功: %s | 大小: %d bytes", path, size)
	return nil
}
