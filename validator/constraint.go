package validator

import (
	"errors"
	"fmt"
	"golang.org/x/exp/constraints"
	"reflect"
	"regexp"
	"time"
	"unicode/utf8"
)

//
// @Author yfy2001
// @Date 2025/7/10 11 19
//

// RangeConstraint 数值范围约束，支持所有有序类型
type RangeConstraint[T constraints.Ordered] struct {
	Min *T // 可选最小值
	Max *T // 可选最大值
}

func (c *RangeConstraint[T]) Validate(value any) error {
	val, ok := value.(T)
	if !ok {
		return fmt.Errorf("类型错误，需要 %T，实际为 %T", *new(T), value)
	}

	if c.Min != nil && val < *c.Min {
		return fmt.Errorf("值 %v 超出范围 (最小值: %v)", val, *c.Min)
	}
	if c.Max != nil && val > *c.Max {
		return fmt.Errorf("值 %v 超出范围 (最大值: %v)", val, *c.Max)
	}
	return nil
}

// LengthConstraint 长度约束，支持多种类型
type LengthConstraint struct {
	Min *int // 最小长度
	Max *int // 最大长度
}

func (c *LengthConstraint) getLength(value any) (int, error) {
	switch v := value.(type) {
	case string:
		return utf8.RuneCountInString(v), nil
	case []byte:
		return len(v), nil
	default:
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan, reflect.String:
			return rv.Len(), nil
		default:
			return 0, fmt.Errorf("不支持长度验证的类型: %T", value)
		}
	}
}

func (c *LengthConstraint) Validate(value any) error {
	length, err := c.getLength(value)
	if err != nil {
		return err
	}

	if c.Min != nil && length < *c.Min {
		return fmt.Errorf("长度 %d 小于最小值 %d", length, *c.Min)
	}
	if c.Max != nil && length > *c.Max {
		return fmt.Errorf("长度 %d 大于最大值 %d", length, *c.Max)
	}
	return nil
}

// EnumConstraint 枚举约束
type EnumConstraint[T comparable] struct {
	Allowed map[T]struct{} // 允许的值集合
}

func (c *EnumConstraint[T]) Validate(value interface{}) error {
	val, ok := value.(T)
	if !ok {
		return fmt.Errorf("类型错误，需要 %T，实际为 %T", *new(T), value)
	}

	if len(c.Allowed) == 0 {
		return fmt.Errorf("非有效的枚举类型")
	}

	if _, exists := c.Allowed[val]; !exists {
		return fmt.Errorf("值 %v 不在枚举范围内", val)
	}
	return nil
}

func NewEnumConstraint[T comparable](values ...T) *EnumConstraint[T] {
	allowed := make(map[T]struct{}, len(values))
	for _, v := range values {
		allowed[v] = struct{}{}
	}
	return &EnumConstraint[T]{Allowed: allowed}
}

// RequiredConstraint 必填约束
type RequiredConstraint struct{}

func (c *RequiredConstraint) Validate(value any) error {
	if value == nil {
		return fmt.Errorf("值不能为nil")
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return fmt.Errorf("字符串不能为空")
		}
	case []byte:
		if len(v) == 0 {
			return fmt.Errorf("字节切片不能为空")
		}
	default:
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
			if rv.IsNil() {
				return fmt.Errorf("值不能为nil")
			}
		default:
			// 对于其他类型，如果它们是“零值”概念上的空，也可以在此处扩展
			// 例如，对于数值类型，0可以认为是空，但这通常取决于具体业务场景
			// 目前的定义主要针对引用类型和字符串/字节切片
			return nil // 对于非引用类型且非字符串/字节切片，不进行“空”检查
		}
	}
	return nil
}

// PatternConstraint 正则表达式约束
type PatternConstraint struct {
	Pattern string
	regex   *regexp.Regexp // 预编译的正则表达式
}

func NewPatternConstraint(pattern string) (*PatternConstraint, error) {
	if pattern == "" {
		return nil, fmt.Errorf("正则表达式不能为空")
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("无效的正则表达式: %w", err)
	}
	return &PatternConstraint{Pattern: pattern, regex: re}, nil
}

func (c *PatternConstraint) Validate(value any) error {
	switch v := value.(type) {
	case string:
		if !c.regex.MatchString(v) {
			return fmt.Errorf("值 %q 不符合格式要求 %q", v, c.Pattern)
		}
	case []byte:
		if !c.regex.Match(v) {
			return fmt.Errorf("字节内容不符合格式要求 %q", c.Pattern)
		}
	default:
		if stringer, ok := value.(fmt.Stringer); ok {
			str := stringer.String()
			if !c.regex.MatchString(str) {
				return fmt.Errorf("值 %q 不符合格式要求 %q", str, c.Pattern)
			}
		} else {
			return fmt.Errorf("需要字符串、字节切片或Stringer类型，实际为 %T", value)
		}
	}
	return nil
}

// TimeConstraint 时间约束
type TimeConstraint struct {
	Format *string // 可选的时间格式（如 "2006-01-02 15:04:05"）
}

func (v *TimeConstraint) Validate(value interface{}) error {
	// 如果有格式要求，验证时间字符串
	if v.Format != nil {
		strVal, ok := value.(string)
		if !ok {
			return errors.New("应为时间字符串")
		}

		_, err := time.Parse(*v.Format, strVal)
		if err != nil {
			return fmt.Errorf("时间格式不符合要求，需要格式 %q", *v.Format)
		}
		return nil
	}

	// 否则验证时间戳
	switch val := value.(type) {
	case int:
		t := time.Unix(int64(val), 0)
		if t.IsZero() {
			return errors.New("无效的时间戳")
		}
	case int64:
		t := time.Unix(val, 0)
		if t.IsZero() {
			return errors.New("无效的时间戳")
		}
	default:
		return errors.New("应为时间戳（整数）或时间字符串")
	}
	return nil
}

type ArrayConstraint struct {
	Item Validator // 每个元素的验证器
}

func (v *ArrayConstraint) Validate(value any) error {
	if v.Item == nil {
		return fmt.Errorf("数组元素验证器不能为空")
	}
	valSlice, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("应为数组类型 []interface{}，实际为 %T", value)
	}

	for i, item := range valSlice {
		if err := v.Item.Validate(item); err != nil {
			return fmt.Errorf("[%d]%v", i, err)
		}
	}
	return nil
}

// TypeConstraint 类型约束验证器
type TypeConstraint struct {
	T reflect.Type // 期望的类型
}

func NewTypeConstraint(t reflect.Type) *TypeConstraint {
	return &TypeConstraint{T: t}
}

func (c *TypeConstraint) Validate(value interface{}) error {
	actual := reflect.TypeOf(value)
	if actual == nil { // 处理 value 是 nil 的情况
		return fmt.Errorf("类型错误，期望 %v，实际为 nil", c.T)
	}

	// 考虑实际类型是否可以赋值给期望类型，而不是严格相等
	if !actual.AssignableTo(c.T) {
		return fmt.Errorf("类型错误，期望 %v，实际为 %v", c.T, actual)
	}
	return nil
}

// CompositeValidator 组合验证器
type CompositeValidator struct {
	validators []Validator
}

func NewCompositeValidator(validators ...Validator) *CompositeValidator {
	return &CompositeValidator{validators: validators}
}

func (cv *CompositeValidator) Validate(value any) error {
	for _, v := range cv.validators {
		if err := v.Validate(value); err != nil {
			// 返回的错误包含是哪个验证器失败以及具体的失败信息
			// %w 用于包装原始错误，方便后续链式处理
			return fmt.Errorf("%w", err)
		}
	}
	return nil
}

// Validate 全局验证函数
// 这个函数更多是作为一种便捷的工具函数，直接传入要验证的值和一系列验证器
func Validate(value any, validators ...Validator) error {
	for _, v := range validators {
		if err := v.Validate(value); err != nil {
			return err
		}
	}
	return nil
}
