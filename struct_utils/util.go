package struct_utils

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"reflect"
)

//
// @Author yfy2001
// @Date 2025/2/28 10 20
//

func ConvertStruct(src, dst interface{}) error {
	srcValue := reflect.ValueOf(src)
	dstValue := reflect.ValueOf(dst)

	// 检查输入是否为指针
	if srcValue.Kind() != reflect.Ptr || dstValue.Kind() != reflect.Ptr {
		return errors.New("both arguments must be pointers to structs")
	}

	srcElem := srcValue.Elem()
	dstElem := dstValue.Elem()

	// 确保元素都是结构体
	if srcElem.Kind() != reflect.Struct || dstElem.Kind() != reflect.Struct {
		return errors.New("arguments must be pointers to structs")
	}

	srcType := srcElem.Type()
	dstType := dstElem.Type()

	// 构建目标结构体字段名到索引的映射
	dstFields := make(map[string]int, dstType.NumField())
	for i := 0; i < dstType.NumField(); i++ {
		dstFields[dstType.Field(i).Name] = i
	}

	// 遍历源结构体的所有字段
	for i := 0; i < srcType.NumField(); i++ {
		srcField := srcElem.Field(i)
		srcFieldType := srcType.Field(i)
		fieldName := srcFieldType.Name

		dstIndex, ok := dstFields[fieldName]
		if !ok {
			continue // 目标结构体无该字段
		}

		dstField := dstElem.Field(dstIndex)

		// 检查目标字段是否可设置
		if !dstField.CanSet() {
			continue
		}

		// 检查类型是否匹配
		if srcField.Type() != dstField.Type() {
			continue
		}

		// 根据字段类型处理
		switch srcField.Kind() {
		case reflect.Struct:
			// 递归处理结构体字段
			if err := ConvertStruct(srcField.Addr().Interface(), dstField.Addr().Interface()); err != nil {
				return err
			}
		case reflect.Ptr:
			srcElemType := srcField.Type().Elem()
			if srcElemType.Kind() == reflect.Struct {
				// 处理结构体指针
				if srcField.IsNil() {
					dstField.Set(reflect.Zero(dstField.Type()))
				} else {
					if dstField.IsNil() {
						// 初始化目标指针
						dstField.Set(reflect.New(srcElemType))
					}
					// 递归处理指针指向的结构体
					if err := ConvertStruct(srcField.Interface(), dstField.Interface()); err != nil {
						return err
					}
				}
			} else {
				// 非结构体指针，直接复制值
				dstField.Set(srcField)
			}
		default:
			// 基础类型直接赋值
			dstField.Set(srcField)
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
	XML
	Gob
)

// UnmarshalData 将数据反序列化为指定类型的结构体
func UnmarshalData[T any](data []byte, format DataFormat) (*T, error) {
	var result T
	var err error

	switch format {
	case JSON:
		err = json.Unmarshal(data, &result)
	case YAML:
		err = yaml.Unmarshal(data, &result)
	case XML:
		err = xml.Unmarshal(data, &result)
	case Gob:
		decoder := gob.NewDecoder(bytes.NewReader(data))
		err = decoder.Decode(&result)
	default:
		return nil, fmt.Errorf("unsupported data format: %v", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %v", err)
	}

	return &result, nil
}

// MarshalData 将结构体序列化为指定格式的数据
func MarshalData(value any, format DataFormat) ([]byte, error) {
	var result []byte
	var err error

	switch format {
	case JSON:
		result, err = json.Marshal(value)
	case YAML:
		result, err = yaml.Marshal(value)
	case XML:
		result, err = xml.Marshal(value)
	case Gob:
		var buffer bytes.Buffer
		encoder := gob.NewEncoder(&buffer)
		err = encoder.Encode(value)
		result = buffer.Bytes()
	default:
		return nil, fmt.Errorf("unsupported data format: %v", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %v", err)
	}

	return result, nil
}
