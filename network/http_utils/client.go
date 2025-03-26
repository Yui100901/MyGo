package http_utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

//
// @Author yfy2001
// @Date 2024/9/27 12 35
//

type HTTPClient struct {
	Client *http.Client
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		Client: &http.Client{},
	}
}

func (c *HTTPClient) SetProxy(proxyFunc func(req *http.Request) (*url.URL, error)) {
	c.Client.Transport = &http.Transport{
		Proxy: proxyFunc,
	}
}

func (c *HTTPClient) SetTimeout(timeout time.Duration) {
	c.Client.Timeout = timeout
}

// GetResponseData 发送HTTP请求读取并返回响应数据
func (c *HTTPClient) GetResponseData(r *HTTPRequest) ([]byte, error) {
	resp, err := c.ExecuteRequest(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应数据
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应错误: %w", err)
	}
	return respData, nil
}

// SaveResponseToFile 发送HTTP请求并将响应数据保存为文件
func (c *HTTPClient) SaveResponseToFile(r *HTTPRequest, filepath string) error {
	resp, err := c.ExecuteRequest(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 创建目标文件
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer out.Close()

	// 将响应数据写入文件
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	return nil
}

// ExecuteRequest 发送HTTP请求并返回响应
func (c *HTTPClient) ExecuteRequest(r *HTTPRequest) (*http.Response, error) {
	req, err := r.generateRequest()
	if err != nil {
		return nil, fmt.Errorf("生成请求失败: %w", err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}

	return resp, nil
}
