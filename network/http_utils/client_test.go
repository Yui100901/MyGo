package http_utils

import (
	"testing"
)

//
// @Author yfy2001
// @Date 2025/1/20 12 50
//

func TestHTTPClient(t *testing.T) {
	c := NewHTTPClient()
	res, err := c.
		Get("http://www.example.com?a=b", nil, nil).
		GetBodyString()
	if err != nil {
		t.Log(err)
	}
	t.Log(res)
}
