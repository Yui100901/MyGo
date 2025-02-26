package http_utils

import (
	"net/http"
	"testing"
)

//
// @Author yfy2001
// @Date 2025/1/20 12 50
//

func TestHTTPClient(t *testing.T) {
	c := NewHTTPClient()
	data, err := c.SendRequest(NewHTTPRequest(http.MethodGet, "http://www.example.com", map[string]string{
		"foo": "bar",
	}, nil, "", nil))
	if err != nil {
		t.Log(err)
	}
	t.Log(string(data))
}
