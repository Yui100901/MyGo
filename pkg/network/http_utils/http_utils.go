package http_utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Yui100901/MyGo/pkg/log_utils"
	"io"
	"net/http"
	"net/url"
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

// HTTPRequest 包含apiUrl和headers
type HTTPRequest struct {
	Method  string
	APIUrl  string
	Headers map[string]string
	ReqData any
}

func NewHTTPRequest(method, apiUrl string, headers map[string]string, reqData any) *HTTPRequest {
	return &HTTPRequest{
		Method:  method,
		APIUrl:  apiUrl,
		Headers: headers,
		ReqData: reqData,
	}
}

func (c *HTTPClient) SendRequest(r *HTTPRequest) ([]byte, error) {
	switch r.Method {
	case "GetByQuery":
		return c.GetByQuery(r)
	case "PostByJson":
		return c.PostByJson(r)
	case "PostByForm":
		return c.PostByForm(r)
	default:
		return nil, fmt.Errorf("invalid method: %s", r.Method)
	}
}

// GetByQuery 发送一个HTTP GET请求到指定的URL，并附带查询参数
func (c *HTTPClient) GetByQuery(r *HTTPRequest) ([]byte, error) {
	// 解析URL并添加查询参数
	reqUrl, err := url.Parse(r.APIUrl)
	if err != nil {
		log_utils.Error.Println("解析URL错误:", err)
		return nil, err
	}

	// 检查请求数据是否为空
	if r.ReqData != nil {
		if queryParams, ok := r.ReqData.(map[string]string); ok {
			query := reqUrl.Query()
			for key, value := range queryParams {
				query.Set(key, value)
			}
			reqUrl.RawQuery = query.Encode()
		} else {
			return nil, fmt.Errorf("请求数据类型错误，预期是 map[string]string")
		}
	}

	// 创建一个新的GET请求
	req, err := http.NewRequest(http.MethodGet, reqUrl.String(), nil)
	if err != nil {
		log_utils.Error.Println("创建请求错误:", err)
		return nil, err
	}

	setHeaders(r.Headers, req)
	// 发送请求并读取响应数据
	return c.doRequest(req)
}

// PostByJson 发送一个带有JSON数据的HTTP POST请求到指定的URL
func (c *HTTPClient) PostByJson(r *HTTPRequest) ([]byte, error) {
	var requestBody []byte
	var err error

	// 检查请求数据是否为空
	if r.ReqData != nil {
		// 将请求数据序列化为JSON
		requestBody, err = json.Marshal(r.ReqData)
		if err != nil {
			log_utils.Error.Println("序列化错误:", err)
			return nil, err
		}
	}

	// 创建一个带有JSON数据的POST请求
	req, err := http.NewRequest(http.MethodPost, r.APIUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		log_utils.Error.Println("创建请求错误:", err)
		return nil, err
	}
	r.Headers["Content-Type"] = "application/json"
	setHeaders(r.Headers, req)
	// 发送请求并读取响应数据
	return c.doRequest(req)
}

// PostByForm 发送一个带有表单数据的HTTP POST请求到指定的URL
func (c *HTTPClient) PostByForm(r *HTTPRequest) ([]byte, error) {
	var requestBody []byte

	// 检查请求数据是否为空
	if r.ReqData != nil {
		if formDataMap, ok := r.ReqData.(map[string]string); ok {
			// 将请求数据编码为表单数据
			formData := url.Values{}
			for key, value := range formDataMap {
				formData.Set(key, value)
			}
			requestBody = []byte(formData.Encode())
		} else {
			return nil, fmt.Errorf("请求数据类型错误，预期是 map[string]string")
		}
	}

	// 创建一个带有表单数据的POST请求
	req, err := http.NewRequest(http.MethodPost, r.APIUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		log_utils.Error.Println("创建请求错误:", err)
		return nil, err
	}
	r.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	setHeaders(r.Headers, req)
	// 发送请求并读取响应数据
	return c.doRequest(req)
}

func setHeaders(h map[string]string, req *http.Request) {
	// 设置请求头
	for key, value := range h {
		req.Header.Set(key, value)
	}
}

// doRequest 发送HTTP请求并读取响应数据，同时设置请求头
func (c *HTTPClient) doRequest(req *http.Request) ([]byte, error) {

	resp, err := c.Client.Do(req)
	if err != nil {
		log_utils.Error.Println("发送请求错误:", err)
		return nil, err
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		log_utils.Error.Println("读取响应错误:", err)
		return nil, err
	}
	return respData, nil
}
