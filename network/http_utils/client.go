package http_utils

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
)

// HTTPClient HTTP客户端封装
type HTTPClient struct {
	Client *http.Client
	Logger *log.Logger // 日志记录器
}

// NewHTTPClient 创建新的HTTP客户端
func NewHTTPClient() *HTTPClient {
	// 创建带 cookie 支持的客户端
	jar, _ := cookiejar.New(nil)

	return &HTTPClient{
		Client: &http.Client{
			Timeout: defaultTimeout,
			Jar:     jar,
		},
		Logger: log.New(os.Stderr, "[HTTP] ", log.LstdFlags),
	}
}

// SetProxy 设置代理
func (c *HTTPClient) SetProxy(proxyURL string) error {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}

	if c.Client.Transport == nil {
		c.Client.Transport = &http.Transport{
			Proxy: http.ProxyURL(parsedURL),
		}
	} else {
		if transport, ok := c.Client.Transport.(*http.Transport); ok {
			transport.Proxy = http.ProxyURL(parsedURL)
		} else {
			return errors.New("cannot set proxy on existing transport")
		}
	}
	return nil
}

// SetTimeout 设置超时时间
func (c *HTTPClient) SetTimeout(timeout time.Duration) {
	c.Client.Timeout = timeout
}

// SetTransport 设置自定义传输层
func (c *HTTPClient) SetTransport(transport *http.Transport) {
	c.Client.Transport = transport
}

// SetInsecureSkipVerify 设置跳过TLS证书验证
func (c *HTTPClient) SetInsecureSkipVerify(skip bool) {
	if c.Client.Transport == nil {
		c.Client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skip},
		}
	} else {
		if transport, ok := c.Client.Transport.(*http.Transport); ok {
			if transport.TLSClientConfig == nil {
				transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: skip}
			} else {
				transport.TLSClientConfig.InsecureSkipVerify = skip
			}
		}
	}
}

// SetLogger 设置自定义日志器
func (c *HTTPClient) SetLogger(logger *log.Logger) {
	c.Logger = logger
}

// ExecuteRequest 执行HTTP请求
func (c *HTTPClient) ExecuteRequest(r *HTTPRequest) (*http.Response, error) {
	req, err := r.generateRequest()
	if err != nil {
		return nil, fmt.Errorf("生成请求失败: %w", err)
	}

	// 记录请求开始时间
	start := time.Now()
	c.Logger.Printf("开始请求: %s %s", req.Method, req.URL)

	resp, err := c.Client.Do(req)
	if err != nil {
		c.Logger.Printf("请求失败: %s %s | 错误: %v | 耗时: %v",
			req.Method, req.URL, err, time.Since(start))
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}

	c.Logger.Printf("请求完成: %s %s | 状态: %d | 耗时: %v",
		req.Method, req.URL, resp.StatusCode, time.Since(start))

	return resp, nil
}

// GetResponseData 发送HTTP请求并返回响应数据
func (c *HTTPClient) GetResponseData(r *HTTPRequest) ([]byte, error) {
	resp, err := c.ExecuteRequest(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("非成功状态码: %d %s", resp.StatusCode, resp.Status)
	}

	// 读取响应数据
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应错误: %w", err)
	}

	c.Logger.Printf("读取响应数据: %s %s | 大小: %d bytes",
		r.Method, r.Url, len(respData))

	return respData, nil
}

// SaveResponseToFile 发送HTTP请求并将响应保存到文件
func (c *HTTPClient) SaveResponseToFile(r *HTTPRequest, filepath string) error {
	resp, err := c.ExecuteRequest(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("非成功状态码: %d %s", resp.StatusCode, resp.Status)
	}

	// 创建目标文件
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 写入数据
	size, err := io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	c.Logger.Printf("文件保存成功: %s | 大小: %d bytes", filepath, size)
	return nil
}
