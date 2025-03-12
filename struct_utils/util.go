package struct_utils

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"reflect"
)

//
// @Author yfy2001
// @Date 2025/2/28 10 20
//

// CopyProperties 使用反射复制 src 的属性值到 dest，包括嵌套结构体
func CopyProperties(src, dest interface{}) error {
	srcVal := reflect.ValueOf(src)
	destVal := reflect.ValueOf(dest)

	if srcVal.Kind() != reflect.Ptr || destVal.Kind() != reflect.Ptr {
		return fmt.Errorf("both src and dest should be pointers")
	}

	srcElem := srcVal.Elem()
	destElem := destVal.Elem()

	if srcElem.Kind() != reflect.Struct || destElem.Kind() != reflect.Struct {
		return fmt.Errorf("src and dest should point to structs")
	}

	for i := 0; i < srcElem.NumField(); i++ {
		srcField := srcElem.Field(i)
		srcFieldName := srcElem.Type().Field(i).Name

		destField := destElem.FieldByName(srcFieldName)
		if !destField.IsValid() || !destField.CanSet() {
			continue
		}

		// 如果字段是结构体，递归复制
		if srcField.Kind() == reflect.Struct && destField.Kind() == reflect.Struct {
			err := CopyProperties(srcField.Addr().Interface(), destField.Addr().Interface())
			if err != nil {
				return err
			}
		} else if srcField.Type() == destField.Type() {
			// 普通字段直接赋值
			destField.Set(srcField)
		}
	}

	return nil
}

// StructToMap 将结构体转换为 map[string]any
func StructToMap(input any) (map[string]any, error) {
	result := make(map[string]any)

	val := reflect.ValueOf(input)
	typ := reflect.TypeOf(input)

	// 必须是结构体或结构体指针
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a struct or a pointer to a struct")
	}

	// 遍历结构体字段
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Name

		// 只处理导出的字段
		if field.CanInterface() {
			result[fieldName] = field.Interface()
		}
	}

	return result, nil
}

// MapToStruct 将 map 转换为结构体
func MapToStruct(data map[string]any, result any) error {
	val := reflect.ValueOf(result)

	// 确保 result 是指向结构体的指针
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("result must be a pointer to a struct")
	}

	valElem := val.Elem()

	for key, value := range data {
		field := valElem.FieldByName(key)
		// 确保字段存在且可设置
		if field.IsValid() && field.CanSet() {
			valValue := reflect.ValueOf(value)
			// 确保类型匹配后进行设置
			if field.Type() == valValue.Type() {
				field.Set(valValue)
			}
		}
	}

	return nil
}

// DataFormat 是一个自定义类型，用于表示数据格式
type DataFormat int

const (
	JSON DataFormat = iota
	YAML
)

// NewStructFromData 根据传入的类型和数据格式进行转换
func NewStructFromData[T any](data []byte, format DataFormat) (*T, error) {
	var result T
	var err error

	switch format {
	case JSON:
		err = json.Unmarshal(data, &result)
	case YAML:
		err = yaml.Unmarshal(data, &result)
	default:
		return nil, fmt.Errorf("unsupported data format")
	}

	if err != nil {
		return nil, fmt.Errorf("error unmarshalling data: %v", err)
	}

	return &result, nil
}
