package converter

import (
	"testing"
)

//
// @Author yfy2001
// @Date 2025/3/18 10 10
//

// TestCapitalize 测试 Capitalize 方法
func TestCapitalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "Hello"},
		{"world", "World"},
		{"", ""},
		{"a", "A"},
	}

	for _, test := range tests {
		output := Capitalize(test.input)
		if output != test.expected {
			t.Errorf("Capitalize(%q) = %q; expected %q", test.input, output, test.expected)
		}
	}
}

// TestCamelToSnake 测试 CamelToSnake 方法
func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"CamelCase", "camel_case"},
		{"SnakeToCamel", "snake_to_camel"},
		{"", ""},
		{"lowercase", "lowercase"},
	}

	for _, test := range tests {
		output := CamelToSnake(test.input)
		if output != test.expected {
			t.Errorf("CamelToSnake(%q) = %q; expected %q", test.input, output, test.expected)
		}
	}
}

// TestSnakeToCamel 测试 SnakeToCamel 方法
func TestSnakeToCamel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"snake_to_camel", "snakeToCamel"},
		{"convert_case", "convertCase"},
		{"", ""},
		{"singleword", "singleword"},
	}

	for _, test := range tests {
		output := SnakeToCamel(test.input)
		if output != test.expected {
			t.Errorf("SnakeToCamel(%q) = %q; expected %q", test.input, output, test.expected)
		}
	}
}
