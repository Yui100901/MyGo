package main

import (
	"bytes"
	"encoding/json"
	"github.com/Yui100901/MyGo/pkg/log_utils"
	"io"
	"net/http"
	"net/url"
)

//
// @Author yfy2001
// @Date 2024/9/27 12 35
//

var client = &http.Client{}

// GetByQuery 发送一个HTTP GET请求到指定的URL，并附带查询参数
func GetByQuery(apiUrl string, headers map[string]string, reqData map[string]string) ([]byte, error) {
	// 解析URL并添加查询参数
	reqUrl, err := url.Parse(apiUrl)
	if err != nil {
		log_utils.Error.Println("解析URL错误:", err)
		return nil, err
	}

	// 检查请求数据是否为空
	if reqData != nil {
		query := reqUrl.Query()
		for key, value := range reqData {
			query.Set(key, value)
		}
		reqUrl.RawQuery = query.Encode()
	}

	// 创建一个新的GET请求
	req, err := http.NewRequest(http.MethodGet, reqUrl.String(), nil)
	if err != nil {
		log_utils.Error.Println("创建请求错误:", err)
		return nil, err
	}

	// 发送请求并读取响应数据
	return doRequest(headers, req)
}

// PostByJson 发送一个带有JSON数据的HTTP POST请求到指定的URL
func PostByJson(apiUrl string, headers map[string]string, reqData any) ([]byte, error) {
	var requestBody []byte
	var err error

	// 检查请求数据是否为空
	if reqData != nil {
		// 将请求数据序列化为JSON
		requestBody, err = json.Marshal(reqData)
		if err != nil {
			log_utils.Error.Println("序列化错误:", err)
			return nil, err
		}
	}

	// 创建一个带有JSON数据的POST请求
	req, err := http.NewRequest(http.MethodPost, apiUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		log_utils.Error.Println("创建请求错误:", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求并读取响应数据
	return doRequest(headers, req)
}

// PostByForm 发送一个带有表单数据的HTTP POST请求到指定的URL
func PostByForm(apiUrl string, headers map[string]string, reqData map[string]string) ([]byte, error) {
	var requestBody []byte

	// 检查请求数据是否为空
	if reqData != nil {
		// 将请求数据编码为表单数据
		formData := url.Values{}
		for key, value := range reqData {
			formData.Set(key, value)
		}
		requestBody = []byte(formData.Encode())
	}

	// 创建一个带有表单数据的POST请求
	req, err := http.NewRequest(http.MethodPost, apiUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		log_utils.Error.Println("创建请求错误:", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送请求并读取响应数据
	return doRequest(headers, req)
}

// doRequest 发送HTTP请求并读取响应数据，同时设置请求头
func doRequest(headers map[string]string, req *http.Request) ([]byte, error) {
	// 设置请求头
	if headers != nil {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	resp, err := client.Do(req)
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
