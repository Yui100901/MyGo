package http_utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//
// @Author yfy2001
// @Date 2025/3/26 08 38
//

// HTTPRequest 结构体封装 HTTP 请求参数
type HTTPRequest struct {
	Method      string            // HTTP 方法 (GET, POST, PUT, DELETE 等)
	URL         string            // 请求 URL
	QueryParams map[string]string // URL 查询参数
	Headers     map[string]string // 请求头
	Body        interface{}       // 请求体 (支持多种类型)
	FormData    map[string]string // 表单数据 (字段名->值)
	FormFiles   map[string]string // 表单文件 (字段名->文件路径)
	Timeout     time.Duration     // 请求超时时间
	Context     context.Context   // 请求上下文
}

// NewHTTPRequest 创建新的 HTTP 请求对象
func NewHTTPRequest(method, url string) *HTTPRequest {
	return &HTTPRequest{
		Method:      method,
		URL:         url,
		QueryParams: make(map[string]string),
		Headers:     make(map[string]string),
		FormData:    make(map[string]string),
		Timeout:     30 * time.Second,
		Context:     context.Background(),
	}
}

// AddQueryParam 添加查询参数
func (r *HTTPRequest) AddQueryParam(key, value string) *HTTPRequest {
	r.QueryParams[key] = value
	Logger.Printf("添加查询参数: %s=%s", key, value)
	return r
}

// SetQueryParams 批量设置查询参数
func (r *HTTPRequest) SetQueryParams(params map[string]string) *HTTPRequest {
	for k, v := range params {
		r.QueryParams[k] = v
	}
	Logger.Printf("设置查询参数: %d 个参数", len(params))
	return r
}

// AddHeader 添加请求头
func (r *HTTPRequest) AddHeader(key, value string) *HTTPRequest {
	r.Headers[key] = value
	Logger.Printf("添加请求头: %s=%s", key, value)
	return r
}

// SetHeaders 批量设置请求头
func (r *HTTPRequest) SetHeaders(headers map[string]string) *HTTPRequest {
	for k, v := range headers {
		r.Headers[k] = v
	}
	Logger.Printf("设置请求头: %d 个请求头", len(headers))
	return r
}

// SetJSONBody 设置 JSON 请求体
func (r *HTTPRequest) SetJSONBody(body interface{}) *HTTPRequest {
	r.Body = body
	r.AddHeader("Content-Type", "application/json")

	// 记录 JSON 摘要
	if jsonData, err := json.Marshal(body); err == nil {
		Logger.Printf("设置JSON请求体: %d", len(jsonData))
	} else {
		Logger.Printf("设置JSON请求体: [无法序列化]")
	}
	return r
}

// AddFormData 添加表单数据
func (r *HTTPRequest) AddFormData(key, value string) *HTTPRequest {
	r.FormData[key] = value
	Logger.Printf("添加表单数据: %s=%s", key, value)
	return r
}

// SetFormData 批量设置表单数据
func (r *HTTPRequest) SetFormData(formData map[string]string) *HTTPRequest {
	r.FormData = formData
	r.AddHeader("Content-Type", "application/x-www-form-urlencoded")
	Logger.Printf("设置表单数据: %d 个字段", len(formData))
	return r
}

// AddFormFile 添加表单文件
func (r *HTTPRequest) AddFormFile(fieldName, filePath string) *HTTPRequest {
	if r.FormFiles == nil {
		r.FormFiles = make(map[string]string)
	}
	r.FormFiles[fieldName] = filePath
	Logger.Printf("添加表单文件: %s -> %s", fieldName, filePath)
	return r
}

// SetTimeout 设置请求超时时间
func (r *HTTPRequest) SetTimeout(timeout time.Duration) *HTTPRequest {
	r.Timeout = timeout
	Logger.Printf("设置超时时间: %v", timeout)
	return r
}

// SetContext 设置请求上下文
func (r *HTTPRequest) SetContext(ctx context.Context) *HTTPRequest {
	r.Context = ctx
	Logger.Printf("设置请求上下文")
	return r
}

// BuildRequest 构建 HTTP 请求
func (r *HTTPRequest) BuildRequest() (*http.Request, error) {
	Logger.Printf("开始构建请求: %s %s", r.Method, r.URL)

	// 处理 URL 和查询参数
	fullURL, err := buildURLWithQuery(r.URL, r.QueryParams)
	if err != nil {
		Logger.Printf("构建URL失败: %v", err)
		return nil, fmt.Errorf("构建URL失败: %w", err)
	}

	Logger.Printf("完整URL: %s", fullURL)

	// 准备请求体
	bodyReader, contentType, err := r.prepareBody()
	if err != nil {
		Logger.Printf("准备请求体失败: %v", err)
		return nil, fmt.Errorf("准备请求体失败: %w", err)
	}

	// 创建请求对象
	req, err := http.NewRequestWithContext(r.Context, r.Method, fullURL, bodyReader)
	if err != nil {
		Logger.Printf("创建请求失败: %v", err)
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	setHeaders(req, r.Headers)

	// 设置内容类型（如果已准备）
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
		Logger.Printf("设置Content-Type: %s", contentType)
	}

	Logger.Printf("请求构建完成")
	return req, nil
}

// buildURLWithQuery 构建带查询参数的URL
func buildURLWithQuery(baseURL string, params map[string]string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	if len(params) > 0 {
		q := u.Query()
		for key, value := range params {
			q.Add(key, value)
		}
		u.RawQuery = q.Encode()
	}

	return u.String(), nil
}

// prepareBody 准备请求体
func (r *HTTPRequest) prepareBody() (io.Reader, string, error) {
	// 如果存在文件，优先处理 multipart/form-data
	if len(r.FormFiles) > 0 {
		return r.prepareMultipartBody()
	}

	// 处理普通请求体
	switch body := r.Body.(type) {
	case nil:
		// 如果没有Body但有FormData，则处理表单
		if len(r.FormData) > 0 {
			return r.prepareFormBody(r.FormData)
		}
		return nil, "", nil
	case []byte:
		Logger.Printf("设置二进制请求体 长度: %d bytes", len(body))
		return bytes.NewReader(body), r.Headers["Content-Type"], nil
	case string:
		Logger.Printf("设置字符串请求体 长度: %d", len(body))
		return strings.NewReader(body), r.Headers["Content-Type"], nil
	case io.Reader:
		Logger.Printf("设置io.Reader请求体")
		return body, r.Headers["Content-Type"], nil
	case map[string]string:
		Logger.Printf("设置map[string]string请求体")
		return r.prepareFormBody(body)
	default:
		// 如果设置了FormData，优先使用FormData
		if len(r.FormData) > 0 {
			return r.prepareFormBody(r.FormData)
		}
		return r.prepareJSONBody()
	}
}

// prepareJSONBody 准备 JSON 请求体
func (r *HTTPRequest) prepareJSONBody() (io.Reader, string, error) {
	jsonData, err := json.Marshal(r.Body)
	if err != nil {
		Logger.Printf("JSON序列化错误: %v", err)
		return nil, "", fmt.Errorf("JSON序列化错误: %w", err)
	}

	Logger.Printf("JSON请求体: %d bytes", len(jsonData))
	return bytes.NewReader(jsonData), "application/json", nil
}

// prepareFormBody 准备表单请求体
func (r *HTTPRequest) prepareFormBody(formData map[string]string) (io.Reader, string, error) {
	values := url.Values{}
	for key, value := range formData {
		values.Add(key, value)
	}

	encoded := values.Encode()
	Logger.Printf("表单请求体: %d bytes", len(encoded))
	return strings.NewReader(encoded), "application/x-www-form-urlencoded", nil
}

// prepareMultipartBody 准备 multipart/form-data 请求体
func (r *HTTPRequest) prepareMultipartBody() (io.Reader, string, error) {
	Logger.Printf("准备multipart/form-data请求体")

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加表单文件
	for fieldName, filePath := range r.FormFiles {
		file, err := os.Open(filePath)
		if err != nil {
			Logger.Printf("打开文件失败: %s - %v", filePath, err)
			return nil, "", fmt.Errorf("打开文件失败: %w", err)
		}

		part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
		if err != nil {
			Logger.Printf("创建表单文件失败: %v", err)
			return nil, "", fmt.Errorf("创建表单文件失败: %w", err)
		}

		if _, err := io.Copy(part, file); err != nil {
			Logger.Printf("写入文件内容失败: %v", err)
			return nil, "", fmt.Errorf("写入文件内容失败: %w", err)
		}

		Logger.Printf("添加文件: %s -> %s (%d bytes)", fieldName, filePath, fileSize(file))
		// 显式关闭文件
		if err := file.Close(); err != nil {
			Logger.Printf("关闭文件失败: %v", err)
		}
	}

	// 添加普通表单字段
	for key, value := range r.FormData {
		if err := writer.WriteField(key, value); err != nil {
			Logger.Printf("写入表单字段失败: %v", err)
			return nil, "", fmt.Errorf("写入表单字段失败: %w", err)
		}
		Logger.Printf("添加表单字段: %s=%s", key, value)
	}

	// 关闭writer以完成multipart写入
	if err := writer.Close(); err != nil {
		Logger.Printf("完成multipart写入失败: %v", err)
		return nil, "", fmt.Errorf("完成multipart写入失败: %w", err)
	}

	contentType := writer.FormDataContentType()
	Logger.Printf("multipart/form-data准备完成: %d bytes", buf.Len())
	return &buf, contentType, nil
}

// setHeaders 设置请求头
func setHeaders(req *http.Request, headers map[string]string) {
	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

// HTTPResponse 封装 HTTP 响应
type HTTPResponse struct {
	StatusCode int               // 状态码
	Headers    map[string]string // 响应头
	Body       []byte            // 响应体
	RequestURL string            // 请求URL
	Duration   time.Duration     // 请求耗时
}

// ParseJSON 解析 JSON 响应
func (resp *HTTPResponse) ParseJSON(target interface{}) error {
	if resp.Body == nil {
		return errors.New("响应体为空")
	}
	return json.Unmarshal(resp.Body, target)
}

// GetBodyString 获取响应体字符串
func (resp *HTTPResponse) GetBodyString() string {
	return string(resp.Body)
}

// IsSuccess 判断请求是否成功 (2xx 状态码)
func (resp *HTTPResponse) IsSuccess() bool {
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
