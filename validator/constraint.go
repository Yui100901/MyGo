package validator

import (
	"errors"
	"fmt"
	"golang.org/x/exp/constraints"
	"reflect"
	"regexp"
	"sync"
	"time"
	"unicode/utf8"
)

//
// @Author yfy2001
// @Date 2025/7/10 11 19
//

// RangeConstraint 数值范围约束，支持所有有序类型（例如：int, float64, string 等）。
type RangeConstraint[T constraints.Ordered] struct {
	Min *T // 可选的最小值指针。如果为 nil，则不检查最小值。
	Max *T // 可选的最大值指针。如果为 nil，则不检查最大值。
}

// Validate 检查给定值是否在指定的数值范围内。
func (c *RangeConstraint[T]) Validate(value interface{}) error {
	val, ok := value.(T)
	if !ok {
		return fmt.Errorf("类型错误：期望类型 %T，实际为 %T", *new(T), value)
	}
	// 使用解引用简化比较逻辑
	rangeMin, rangeMax := c.Min, c.Max
	if rangeMin != nil && val < *rangeMin {
		return fmt.Errorf("值 %v 超出范围：小于最小值 %v", val, *rangeMin)
	}
	if rangeMax != nil && val > *rangeMax {
		return fmt.Errorf("值 %v 超出范围：大于最大值 %v", val, *rangeMax)
	}
	return nil
}

// LengthConstraint 长度约束，支持字符串、字节切片、数组、切片、映射和通道等类型。
type LengthConstraint struct {
	Min *int // 最小长度指针。如果为 nil，则不检查最小长度。
	Max *int // 最大长度指针。如果为 nil，则不检查最大长度。
}

// getLength 根据值的类型获取其长度。
// 添加对指针类型的支持
func (c *LengthConstraint) getLength(value interface{}) (int, error) {
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return 0, nil
		}
		return c.getLength(rv.Elem().Interface())
	case reflect.String:
		return utf8.RuneCountInString(rv.String()), nil
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return rv.Len(), nil
	default:
		return 0, fmt.Errorf("不支持长度验证的类型：%T", value)
	}
}

// Validate 检查给定值的长度是否在指定范围内。
func (c *LengthConstraint) Validate(value interface{}) error {
	length, err := c.getLength(value)
	if err != nil {
		return err // 如果获取长度失败，直接返回错误
	}

	if c.Min != nil && length < *c.Min {
		// 如果设置了最小长度且当前长度小于最小长度，则返回错误。
		return fmt.Errorf("长度 %d 小于最小值 %d", length, *c.Min)
	}
	if c.Max != nil && length > *c.Max {
		// 如果设置了最大长度且当前长度大于最大长度，则返回错误。
		return fmt.Errorf("长度 %d 大于最大值 %d", length, *c.Max)
	}
	return nil // 验证通过
}

// EnumConstraint 枚举约束，用于检查值是否在预定义的允许值集合中。
type EnumConstraint[T comparable] struct {
	Allowed map[T]struct{} // 允许的值集合，使用 map 实现高效查找
}

// NewEnumConstraint 是创建 EnumConstraint 的便捷函数。
func NewEnumConstraint[T comparable](values ...T) *EnumConstraint[T] {
	allowed := make(map[T]struct{}, len(values))
	for _, v := range values {
		allowed[v] = struct{}{} // 将所有允许的值添加到 map 中
	}
	return &EnumConstraint[T]{Allowed: allowed}
}

// Validate 检查给定值是否包含在枚举允许的值集合中。
func (c *EnumConstraint[T]) Validate(value interface{}) error {
	if len(c.Allowed) == 0 {
		// 如果允许值集合为空，则认为是非有效的枚举配置。
		return fmt.Errorf("枚举约束配置无效：允许值集合为空")
	}

	val, ok := value.(T)
	if !ok {
		// 检查类型是否匹配，如果不匹配则返回类型错误。
		return fmt.Errorf("类型错误：期望类型 %T，实际为 %T", *new(T), value)
	}

	if _, exists := c.Allowed[val]; !exists {
		// 如果值不在允许值集合中，则返回错误。
		return fmt.Errorf("值 %v 不在枚举范围内", val)
	}
	return nil // 验证通过
}

// RequiredConstraint 必填约束，用于检查值是否为空（nil、空字符串、空切片等）。
type RequiredConstraint struct{}

// Validate 检查给定值是否为必需的（非空）。
func (c *RequiredConstraint) Validate(value interface{}) error {
	if value == nil {
		return fmt.Errorf("值不能为 nil")
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		return errors.New("指针不能为 nil")
	}
	// 简化空值检查逻辑
	if rv.Kind() == reflect.String && rv.Len() == 0 {
		return errors.New("字符串不能为空")
	}
	if rv.Kind() == reflect.Slice && rv.Len() == 0 {
		return errors.New("切片不能为空")
	}
	return nil
}

// PatternConstraint 正则表达式约束，用于检查字符串或字节切片是否符合指定的正则表达式。
// 在 PatternConstraint 中使用 sync.Once 延迟编译正则
type PatternConstraint struct {
	Pattern string
	once    sync.Once
	regex   *regexp.Regexp
	initErr error
}

func (c *PatternConstraint) compile() {
	c.once.Do(func() {
		if c.Pattern == "" {
			c.initErr = errors.New("正则表达式不能为空")
			return
		}
		c.regex, c.initErr = regexp.Compile(c.Pattern)
	})
}

// NewPatternConstraint 是创建 PatternConstraint 的便捷函数，并预编译正则表达式。
func NewPatternConstraint(pattern string) (*PatternConstraint, error) {
	if pattern == "" {
		return nil, fmt.Errorf("正则表达式不能为空")
	}
	return &PatternConstraint{Pattern: pattern}, nil
}

// Validate 检查给定值是否符合指定的正则表达式模式。
func (c *PatternConstraint) Validate(value interface{}) error {
	c.compile()
	if c.initErr != nil {
		return c.initErr
	}
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v) // 将字节切片转换为字符串进行匹配
	case fmt.Stringer:
		s = v.String() // 如果实现了 fmt.Stringer 接口，则使用其 String() 方法
	default:
		return fmt.Errorf("类型错误：期望字符串、字节切片或实现了 fmt.Stringer 接口的类型，实际为 %T", value)
	}

	if !c.regex.MatchString(s) {
		// 如果字符串不匹配正则表达式，则返回错误。
		return fmt.Errorf("值 %q 不符合格式要求 %q", s, c.Pattern)
	}
	return nil // 验证通过
}

// TimeConstraint 时间约束，用于验证时间戳或时间字符串。
type TimeConstraint struct {
	Format *string    // 可选的时间格式字符串（例如 "2006-01-02 15:04:05"）。如果为 nil，则验证时间戳。
	Min    *time.Time // 可选的最小时间点。
	Max    *time.Time // 可选的最大时间点。
}

// Validate 检查给定值是否为有效的时间戳或符合指定格式的时间字符串，并可选地检查时间范围。
func (v *TimeConstraint) Validate(value interface{}) error {
	var t time.Time
	var err error

	if v.Format != nil {
		// 如果指定了格式，则期望是时间字符串。
		strVal, ok := value.(string)
		if !ok {
			return errors.New("类型错误：期望时间字符串")
		}
		t, err = time.Parse(*v.Format, strVal) // 按指定格式解析字符串
		if err != nil {
			return fmt.Errorf("时间格式不符合要求 %q: %w", *v.Format, err)
		}
	} else {
		// 如果未指定格式，则期望是时间戳（int, int64）或 time.Time 类型。
		switch val := value.(type) {
		case int:
			t = time.Unix(int64(val), 0) // 将 int 转换为 Unix 时间戳
		case int64:
			t = time.Unix(val, 0) // 将 int64 转换为 Unix 时间戳
		case time.Time:
			t = val // 直接使用 time.Time 类型
		default:
			return errors.New("类型错误：期望时间戳（整数）或 time.Time 类型")
		}

		// 简单的有效性检查：如果解析后的时间是零值且不是 Unix 纪元，则可能无效。
		// 更严格的检查可能需要判断时间戳是否为负数或超出合理范围。
		if t.IsZero() && value != 0 && value != int64(0) {
			return errors.New("无效的时间值")
		}
	}

	// 检查时间范围
	if v.Min != nil && t.Before(*v.Min) {
		return fmt.Errorf("时间 %s 早于最小值 %s", t.Format(time.RFC3339), v.Min.Format(time.RFC3339))
	}
	if v.Max != nil && t.After(*v.Max) {
		return fmt.Errorf("时间 %s 晚于最大值 %s", t.Format(time.RFC3339), v.Max.Format(time.RFC3339))
	}

	return nil // 验证通过
}

// ArrayConstraint 数组/切片约束，用于验证集合中每个元素的有效性。
type ArrayConstraint struct {
	Item Validator // 每个元素的验证器
}

// Validate 检查给定值是否为数组或切片类型，并对其每个元素应用 Item 验证器。
func (v *ArrayConstraint) Validate(value interface{}) error {
	if v.Item == nil {
		return fmt.Errorf("数组元素验证器不能为空")
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			// 对每个元素进行验证
			if err := v.Item.Validate(rv.Index(i).Interface()); err != nil {
				// 返回带有索引和原始错误的详细错误信息。
				return fmt.Errorf("[%d]%w", i, err)
			}
		}
	default:
		// 如果不是数组或切片类型，则返回错误。
		return fmt.Errorf("类型错误：期望数组或切片类型，实际为 %T", value)
	}
	return nil // 验证通过
}

// TypeConstraint 类型约束验证器，用于检查值的实际类型是否符合期望。
type TypeConstraint struct {
	ExpectedType reflect.Type // 期望的 Go 类型
}

// NewTypeConstraint 是创建 TypeConstraint 的便捷函数。
func NewTypeConstraint(t reflect.Type) *TypeConstraint {
	return &TypeConstraint{ExpectedType: t}
}

// Validate 检查给定值的类型是否可以赋值给期望的类型。
func (c *TypeConstraint) Validate(value interface{}) error {
	actual := reflect.TypeOf(value)
	if actual == nil {
		// 如果值为 nil，且期望的类型不是接口或指针，则返回错误。
		// 这里可以根据实际需求调整 nil 值的处理逻辑。
		if c.ExpectedType.Kind() != reflect.Interface && c.ExpectedType.Kind() != reflect.Ptr {
			return fmt.Errorf("类型错误：期望 %v，实际为 nil", c.ExpectedType)
		}
		return nil // 如果期望的是接口或指针，nil 可能是有效的
	}

	// 使用 AssignableTo 检查实际类型是否可以赋值给期望类型，这比直接相等更灵活。
	if !actual.AssignableTo(c.ExpectedType) {
		return fmt.Errorf("类型错误：期望类型 %v，实际类型 %v", c.ExpectedType, actual)
	}
	return nil // 验证通过
}

// CompositeValidator 组合验证器，允许将多个验证器链式应用于同一个值。
type CompositeValidator struct {
	validators []Validator // 包含的验证器切片
}
