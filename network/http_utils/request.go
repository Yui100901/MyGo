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
)

//
// @Author yfy2001
// @Date 2025/3/26 08 38
//

// HTTPRequest 包含apiUrl和headers
type HTTPRequest struct {
	Method string
	Url    string
	Query  map[string]string
	Header map[string]string
	Body   any
}

func NewHTTPRequest(method, apiUrl string, query, header map[string]string, body any) *HTTPRequest {
	return &HTTPRequest{
		Method: method,
		Url:    apiUrl,
		Query:  query,
		Header: header,
		Body:   body,
	}
}

func (hr *HTTPRequest) generateRequest() (*http.Request, error) {
	req, _ := http.NewRequest(hr.Method, "", nil)
	err := setUrlWithQuery(req, hr.Url, hr.Query)
	if err != nil {
		return nil, err
	}
	if len(hr.Header) != 0 {
		setHeader(req, hr.Header)
	}
	if hr.Body != nil {
		contentType := req.Header.Get("Content-Type")
		err := setBody(req, contentType, hr.Body)
		if err != nil {
			return nil, err
		}
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
	req.URL = reqUrl
	if len(query) != 0 {
		reqUrl.RawQuery = encodeUrlValues(query)
	}
	return nil
}

func setHeader(req *http.Request, headerMap map[string]string) {
	for key, value := range headerMap {
		req.Header.Set(key, value)
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
