package converter

import (
	"strings"
	"unicode"
)

//
// @Author yfy2001
// @Date 2025/3/4 13 36
//

// Capitalize 字符首字母大写
func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

// CamelToSnake 驼峰转下划线
func CamelToSnake(s string) string {
	runes := []rune(s)
	var result []rune
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i != 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// SnakeToCamel 下划线转驼峰
func SnakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if i > 0 {
			parts[i] = Capitalize(parts[i])
		}
	}
	return strings.Join(parts, "")
}
