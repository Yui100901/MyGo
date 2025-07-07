package http_utils

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"
)

// HTTPClient HTTP客户端封装
type HTTPClient struct {
	Client *http.Client
}

// NewHTTPClient 创建新的HTTP客户端
func NewHTTPClient() *HTTPClient {
	// 创建带 cookie 支持的客户端
	jar, _ := cookiejar.New(nil)

	return &HTTPClient{
		Client: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
	}
}

// SetProxy 设置代理
func (c *HTTPClient) SetProxy(proxyURL string) error {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		Logger.Printf("设置代理失败: 无效URL %s - %v", proxyURL, err)
		return fmt.Errorf("无效的代理URL: %w", err)
	}

	if c.Client.Transport == nil {
		c.Client.Transport = &http.Transport{
			Proxy: http.ProxyURL(parsedURL),
		}
	} else {
		if transport, ok := c.Client.Transport.(*http.Transport); ok {
			transport.Proxy = http.ProxyURL(parsedURL)
		} else {
			Logger.Printf("设置代理失败: 传输层类型不匹配")
			return errors.New("无法在现有传输层上设置代理")
		}
	}

	Logger.Printf("设置代理: %s", proxyURL)
	return nil
}

// SetTimeout 设置超时时间
func (c *HTTPClient) SetTimeout(timeout time.Duration) {
	c.Client.Timeout = timeout
	Logger.Printf("设置全局超时: %v", timeout)
}

// SetTransport 设置自定义传输层
func (c *HTTPClient) SetTransport(transport *http.Transport) {
	c.Client.Transport = transport
	Logger.Printf("设置自定义传输层")
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

	Logger.Printf("设置TLS证书验证跳过: %v", skip)
}

// SetCookieJar 设置自定义 Cookie Jar
func (c *HTTPClient) SetCookieJar(jar http.CookieJar) {
	c.Client.Jar = jar
	Logger.Printf("设置自定义Cookie Jar")
}

// EnableConnectionPool 启用连接池
func (c *HTTPClient) EnableConnectionPool(maxIdle, maxIdlePerHost int) {
	if c.Client.Transport == nil {
		c.Client.Transport = &http.Transport{
			MaxIdleConns:        maxIdle,
			MaxIdleConnsPerHost: maxIdlePerHost,
		}
	} else {
		if transport, ok := c.Client.Transport.(*http.Transport); ok {
			transport.MaxIdleConns = maxIdle
			transport.MaxIdleConnsPerHost = maxIdlePerHost
		}
	}

	Logger.Printf("启用连接池: MaxIdleConns=%d, MaxIdleConnsPerHost=%d", maxIdle, maxIdlePerHost)
}

// Do 执行HTTP请求
func (c *HTTPClient) Do(req *http.Request) (*HTTPResponse, error) {
	start := time.Now()
	Logger.Printf("开始请求: %s %s", req.Method, req.URL.String())

	// 记录请求头 (过滤敏感信息)
	c.logRequestHeaders(req)

	resp, err := c.Client.Do(req)
	duration := time.Since(start)

	if err != nil {
		Logger.Printf("请求失败: %s %s | 错误: %v | 耗时: %v",
			req.Method, req.URL, err, duration)
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	Logger.Printf("请求完成: %s %s | 状态: %d | 耗时: %v",
		req.Method, req.URL, resp.StatusCode, duration)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		Logger.Printf("读取响应体失败: %v", err)
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	Logger.Printf("响应体大小: %d bytes", len(body))

	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// 记录响应头 (过滤敏感信息)
	c.logResponseHeaders(headers)

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       body,
		RequestURL: req.URL.String(),
		Duration:   duration,
	}, nil
}

// Execute 执行HTTP请求并返回响应
func (c *HTTPClient) Execute(request *HTTPRequest) (*HTTPResponse, error) {
	req, err := request.BuildRequest()
	if err != nil {
		Logger.Printf("构建请求失败: %v", err)
		return nil, fmt.Errorf("构建请求失败: %w", err)
	}
	return c.Do(req)
}

// Get 执行GET请求
func (c *HTTPClient) Get(url string, headers map[string]string) (*HTTPResponse, error) {
	req := NewHTTPRequest(http.MethodGet, url)
	req.SetHeaders(headers)
	return c.Execute(req)
}

func (c *HTTPClient) Post(url string, headers map[string]string) (*HTTPResponse, error) {
	req := NewHTTPRequest(http.MethodPost, url)
	req.SetHeaders(headers)
	return c.Execute(req)
}

// PostJSON 执行POST JSON请求
func (c *HTTPClient) PostJSON(url string, body interface{}, headers map[string]string) (*HTTPResponse, error) {
	req := NewHTTPRequest(http.MethodPost, url)
	req.SetJSONBody(body)
	req.SetHeaders(headers)
	return c.Execute(req)
}

// PostForm 执行表单POST请求
func (c *HTTPClient) PostForm(url string, formData map[string]string, headers map[string]string) (*HTTPResponse, error) {
	req := NewHTTPRequest(http.MethodPost, url)
	req.SetFormData(formData)
	req.SetHeaders(headers)
	return c.Execute(req)
}

// PostMultipart 执行multipart/form-data POST请求
func (c *HTTPClient) PostMultipart(url string, formData map[string]string, files map[string]string, headers map[string]string) (*HTTPResponse, error) {
	req := NewHTTPRequest(http.MethodPost, url)
	req.FormData = formData
	for field, filePath := range files {
		req.AddFormFile(field, filePath)
	}
	req.SetHeaders(headers)
	return c.Execute(req)
}

// DownloadFile 下载文件到指定路径
func (c *HTTPClient) DownloadFile(url, filePath string) error {
	Logger.Printf("开始下载文件: %s -> %s", url, filePath)

	start := time.Now()
	resp, err := c.Get(url, nil)
	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		err := fmt.Errorf("下载失败: 状态码 %d", resp.StatusCode)
		Logger.Printf("下载失败: %v", err)
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		Logger.Printf("创建文件失败: %v", err)
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	_, err = file.Write(resp.Body)
	if err != nil {
		Logger.Printf("写入文件失败: %v", err)
		return fmt.Errorf("写入文件失败: %w", err)
	}

	Logger.Printf("文件下载成功: %s | 大小: %d bytes | 耗时: %v",
		filePath, len(resp.Body), time.Since(start))
	return nil
}
