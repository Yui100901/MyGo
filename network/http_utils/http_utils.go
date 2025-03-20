package http_utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
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

// HTTPRequest 包含apiUrl和headers
type HTTPRequest struct {
	Method      string
	APIUrl      string
	Query       map[string]string
	Header      map[string]string
	ContentType string
	Body        any
}

func NewHTTPRequest(method, apiUrl string, query, header map[string]string, contentType string, body any) *HTTPRequest {
	return &HTTPRequest{
		Method:      method,
		APIUrl:      apiUrl,
		Query:       query,
		Header:      header,
		ContentType: contentType,
		Body:        body,
	}
}

func (hr *HTTPRequest) generateRequest() (*http.Request, error) {
	req, _ := http.NewRequest(hr.Method, "", nil)
	err := setUrlWithQuery(req, hr.APIUrl, hr.Query)
	if err != nil {
		return nil, err
	}
	if hr.ContentType != "" {
		hr.Header["Content-Type"] = hr.ContentType
	}
	setHeader(req, hr.Header)
	err = setBody(req, hr.ContentType, hr.Body)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func encodeUrlValues(data map[string]string) string {
	values := url.Values{}
	for k, v := range data {
		values.Add(k, v)
	}
	return values.Encode()
}

func setUrlWithQuery(req *http.Request, url1 string, query map[string]string) error {
	reqUrl, err := url.Parse(url1)
	if err != nil {
		return err
	}
	reqUrl.RawQuery = encodeUrlValues(query)
	req.URL = reqUrl
	return nil
}

func setHeader(req *http.Request, headerMap map[string]string) {
	for key, value := range headerMap {
		req.Header.Add(key, value)
	}
}

func setBody(req *http.Request, contentType string, data any) error {
	switch contentType {
	case "application/json":
		jsonData, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("JSON序列化错误: %v", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(jsonData))
	case "application/x-www-form-urlencoded":
		if formDataMap, ok := data.(map[string]string); ok {
			formData := url.Values{}
			for key, value := range formDataMap {
				formData.Set(key, value)
			}
			req.Body = io.NopCloser(strings.NewReader(formData.Encode()))
		} else {
			return fmt.Errorf("请求数据类型错误，预期是 map[string]string")
		}
	case "multipart/form-data":
		if formDataMap, ok := data.(map[string]io.Reader); ok {
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)
			for key, reader := range formDataMap {
				part, err := writer.CreateFormFile(key, key)
				if err != nil {
					return fmt.Errorf("创建multipart/form文件错误: %v", err)
				}
				if _, err = io.Copy(part, reader); err != nil {
					return fmt.Errorf("写入multipart/form文件错误: %v", err)
				}
			}
			writer.Close()
			req.Body = io.NopCloser(&buf)
			req.Header.Set("Content-Type", writer.FormDataContentType())
		} else {
			return fmt.Errorf("请求数据类型错误，预期是 map[string]io.Reader")
		}
	}
	return nil
}

// SendRequest 发送HTTP请求并读取响应数据
func (c *HTTPClient) SendRequest(r *HTTPRequest) ([]byte, error) {

	req, err := r.generateRequest()
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求错误:%w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应错误:%w", err)
	}
	return respData, nil
}
