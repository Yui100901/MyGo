package http_utils

import "testing"

//
// @Author yfy2001
// @Date 2025/1/20 12 50
//

func TestHTTPClient(t *testing.T) {
	c := NewHTTPClient()
	data, _ := c.SendRequest(NewHTTPRequest("GetByQuery", "http://www.example.com", nil, nil))
	t.Log(string(data))
}
