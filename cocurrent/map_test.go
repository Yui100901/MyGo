package cocurrent

import (
	"testing"
)

//
// @Author yfy2001
// @Date 2025/1/19 11 23
//

func TestGenericSyncMap_Keys(t *testing.T) {
	m := NewGenericSyncMap[string, string]()
	m.Set("foo", "bar")
	m.Set("foo1", "bar1")
	keys := m.Keys()
	for _, key := range keys {
		t.Log(key)
	}
}

func TestGenericSyncMap_Values(t *testing.T) {
	m := NewGenericSyncMap[string, string]()
	m.Set("foo", "bar")
	m.Set("foo1", "bar1")
	values := m.Values()
	for _, value := range values {
		t.Log(value)
	}
}

func TestGenericSyncMap_GetOr(t *testing.T) {
	m := NewGenericSyncMap[string, string]()
	m.Set("foo", "bar")
	m.Set("foo1", "bar1")
	v, _ := m.Get("foo")
	t.Log(v)
	v = m.GetOr("b", "b")
	t.Log(v)
	v = m.GetOrElse("a", func() string {
		return "a"
	})
	t.Log(v)
}

func TestSafeMap_GetOr(t *testing.T) {
	m := NewSafeMap[string, string]()
	m.Set("foo", "bar")
	m.Set("foo1", "bar1")
	v, _ := m.Get("foo")
	t.Log(v)
	v = m.GetOr("b", "b")
	t.Log(v)
	v = m.GetOrElse("a", func() string {
		return "a"
	})
	t.Log(v)
}
