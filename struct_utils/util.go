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

// ConvertStruct 递归地将源结构体（src）的字段值复制到目标结构体指针（dest）中。
// src 支持结构体或结构体指针，dest 必须为结构体指针。
func ConvertStruct(src, dest interface{}) error {
	srcVal := reflect.ValueOf(src)
	destVal := reflect.ValueOf(dest)

	// 验证 dest 必须是非 nil 的结构体指针
	if destVal.Kind() != reflect.Ptr || destVal.IsNil() {
		return errors.New("dest must be a non-nil pointer")
	}
	if destVal.Elem().Kind() != reflect.Struct {
		return errors.New("dest must be a pointer to struct")
	}
	destElem := destVal.Elem()

	// 处理 src，允许结构体或结构体指针
	var srcElem reflect.Value
	if srcVal.Kind() == reflect.Ptr {
		if srcVal.IsNil() {
			return errors.New("src pointer is nil")
		}
		srcElem = srcVal.Elem()
	} else {
		srcElem = srcVal
	}
	if srcElem.Kind() != reflect.Struct {
		return errors.New("src must be struct or struct pointer")
	}

	// 构建目标结构体的导出字段名与字段索引映射
	destType := destElem.Type()
	destFields := make(map[string]int, destType.NumField())
	for i := 0; i < destType.NumField(); i++ {
		field := destType.Field(i)
		if field.IsExported() {
			destFields[field.Name] = i
		}
	}

	// 遍历 src 的所有导出字段
	srcType := srcElem.Type()
	for i := 0; i < srcType.NumField(); i++ {
		sf := srcType.Field(i)
		if !sf.IsExported() {
			continue
		}

		fieldName := sf.Name
		dstIdx, exists := destFields[fieldName]
		if !exists {
			continue
		}

		srcField := srcElem.Field(i)
		dstField := destElem.Field(dstIdx)
		if !dstField.CanSet() {
			continue
		}

		// 若目标字段是指针且为 nil，则先初始化字段
		if dstField.Kind() == reflect.Ptr && dstField.IsNil() {
			dstField.Set(reflect.New(dstField.Type().Elem()))
		}

		// 深层复制字段
		if err := copyField(srcField, dstField); err != nil {
			return fmt.Errorf("field %s: %w", fieldName, err)
		}
	}
	return nil
}

// copyField 处理字段的深层复制逻辑，包括指针与结构体之间的转换
func copyField(src, dst reflect.Value) error {
	// 如果 src 是指针且为 nil，则将目标指针置空
	if src.Kind() == reflect.Ptr && src.IsNil() {
		if dst.Kind() == reflect.Ptr {
			dst.Set(reflect.Zero(dst.Type()))
		}
		return nil
	}

	// 类型完全匹配时直接赋值
	if src.Type().AssignableTo(dst.Type()) {
		dst.Set(src)
		return nil
	}

	// 当目标为指针时处理
	if dst.Kind() == reflect.Ptr {
		dstElemType := dst.Type().Elem()

		// 如果 src 也是指针，则先解引用
		if src.Kind() == reflect.Ptr {
			srcElem := src.Elem()
			// 如果解引用后的类型可以直接赋值给目标元素
			if srcElem.Type().AssignableTo(dstElemType) {
				if dst.IsNil() {
					dst.Set(reflect.New(dstElemType))
				}
				dst.Elem().Set(srcElem)
				return nil
			}
			// 如果两者均为结构体，则递归转换（指针结构体→指针结构体的场景也被包含）
			if srcElem.Kind() == reflect.Struct && dstElemType.Kind() == reflect.Struct {
				if dst.IsNil() {
					dst.Set(reflect.New(dstElemType))
				}
				return ConvertStruct(getPointer(srcElem), dst.Elem().Addr().Interface())
			}
		} else { // src 不是指针
			if src.Type().AssignableTo(dstElemType) {
				if dst.IsNil() {
					dst.Set(reflect.New(dstElemType))
				}
				dst.Elem().Set(src)
				return nil
			}
			// 如 src 和目标指针指向的类型均是结构体，则递归转换
			if src.Kind() == reflect.Struct && dstElemType.Kind() == reflect.Struct {
				if dst.IsNil() {
					dst.Set(reflect.New(dstElemType))
				}
				return ConvertStruct(getPointer(src), dst.Elem().Addr().Interface())
			}
		}
	}

	// 当 src 为指针而目标不是指针时，解引用后进行转换
	if src.Kind() == reflect.Ptr {
		srcElem := src.Elem()
		if srcElem.Type().AssignableTo(dst.Type()) {
			dst.Set(srcElem)
			return nil
		}
		if srcElem.Kind() == reflect.Struct && dst.Kind() == reflect.Struct {
			return ConvertStruct(getPointer(srcElem), dst.Addr().Interface())
		}
	}

	// 直接结构体到结构体转换
	if src.Kind() == reflect.Struct && dst.Kind() == reflect.Struct {
		return ConvertStruct(getPointer(src), dst.Addr().Interface())
	}

	// 类型不兼容，忽略该字段
	return nil
}

// getPointer 确保获取传入值的可寻址指针形式，解决不可寻址的问题
func getPointer(v reflect.Value) interface{} {
	if v.CanAddr() {
		return v.Addr().Interface()
	}
	cp := reflect.New(v.Type()).Elem()
	cp.Set(v)
	return cp.Addr().Interface()
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

func ConvertTo[T any](value any) (T, error) {
	var val T
	var ok bool

	// 首先尝试直接类型断言
	val, ok = value.(T)
	if !ok {
		// 类型断言失败：尝试通过反射转换
		v := reflect.ValueOf(value)
		tType := reflect.TypeOf(val) // 获取T的具体类型

		// 检查值类型是否可转换为T
		if !v.Type().ConvertibleTo(tType) {
			return val, fmt.Errorf("can not convert: %T", value)
		}

		// 执行安全转换
		converted := v.Convert(tType)
		val = converted.Interface().(T) // 转换成功后可安全断言
	}
	return val, nil
}
