package validator

import (
	"errors"
	"fmt"
	"github.com/Yui100901/MyGo/log_utils"
	"golang.org/x/exp/constraints"
	"reflect"
	"regexp"
	"sync" // For sync.Once in PatternConstraint
	"time"
	"unicode/utf8"
)

// @Author yfy2001
// @Date 2025/7/10 11:19

// Validator 接口定义了所有验证器必须实现的方法。
// (此接口的定义移至 validator.go 以保持一致性)

func Convert[T any](value any) (T, error) {
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
			return val, fmt.Errorf("类型错误：期望类型 %T，实际为 %T", val, value)
		}

		// 执行安全转换
		converted := v.Convert(tType)
		val = converted.Interface().(T) // 转换成功后可安全断言
	}
	return val, nil
}

// RangeConstraint 数值范围约束，支持所有有序类型（例如：int, float64, string 等）。
type RangeConstraint[T constraints.Ordered] struct {
	Min *T // 可选的最小值指针。如果为 nil，则不检查最小值。
	Max *T // 可选的最大值指针。如果为 nil，则不检查最大值。
}

// Validate 检查给定值是否在指定的数值范围内。
func (c *RangeConstraint[T]) Validate(value interface{}) error {
	val, err := Convert[T](value)
	if err != nil {
		return err
	}

	// 使用局部变量简化比较逻辑
	if c.Min != nil && val < *c.Min {
		return fmt.Errorf("值 %v 超出范围：小于最小值 %v", val, *c.Min)
	}
	if c.Max != nil && val > *c.Max {
		return fmt.Errorf("值 %v 超出范围：大于最大值 %v", val, *c.Max)
	}
	return nil // 验证通过
}

// LengthConstraint 长度约束，支持字符串、字节切片、数组、切片、映射和通道等类型。
type LengthConstraint struct {
	Min *int // 最小长度指针。如果为 nil，则不检查最小长度。
	Max *int // 最大长度指针。如果为 nil，则不检查最大长度。
}

// getLength 根据值的类型获取其长度。
// 增加了对指针类型和接口类型的递归支持。
func (c *LengthConstraint) getLength(value interface{}) (int, error) {
	if value == nil {
		return 0, nil // nil 值的长度视为 0
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		// 递归解引用指针或接口，直到获取到具体值
		if rv.IsNil() {
			return 0, nil
		}
		return c.getLength(rv.Elem().Interface())
	case reflect.String:
		// 字符串按 Unicode 字符数计算长度
		return utf8.RuneCountInString(rv.String()), nil
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		// 数组、切片、映射、通道按其元素数量计算长度
		return rv.Len(), nil
	default:
		// 对于不支持长度验证的类型，返回错误。
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

	val, err := Convert[T](value)
	if err != nil {
		return err
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
// 优化为更全面的 nil 和空值检查。
func (c *RequiredConstraint) Validate(value interface{}) error {
	if value == nil {
		return errors.New("值不能为 nil")
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Func:
		// 对于引用类型，检查其是否为 nil
		if rv.IsNil() {
			return fmt.Errorf("%s 类型不能为 nil", rv.Kind().String())
		}
	case reflect.Slice, reflect.Map, reflect.Chan:
		if rv.IsNil() {
			return fmt.Errorf("%s 类型不能为 nil", rv.Kind().String())
		}
		if rv.Len() == 0 {
			return fmt.Errorf("%s 不能为空", rv.Kind().String())
		}
	case reflect.String, reflect.Array:
		// 对于字符串和数组，检查其长度是否为 0
		log_utils.Info.Println(rv.Type())
		if rv.Len() == 0 {
			return fmt.Errorf("%s 类型不能为空", rv.Kind().String())
		}
	default:
		return nil
	}
	return nil // 验证通过
}

// PatternConstraint 正则表达式约束，用于检查字符串或字节切片是否符合指定的正则表达式。
// 使用 sync.Once 确保正则表达式只编译一次，且在第一次 Validate 调用时进行惰性编译。
type PatternConstraint struct {
	Pattern string
	once    sync.Once
	regex   *regexp.Regexp
	initErr error // 用于存储正则表达式编译时可能发生的错误
}

// compile 编译正则表达式。此方法通过 sync.Once 确保只执行一次。
func (c *PatternConstraint) compile() {
	c.once.Do(func() {
		if c.Pattern == "" {
			c.initErr = errors.New("正则表达式不能为空")
			return
		}
		c.regex, c.initErr = regexp.Compile(c.Pattern)
	})
}

// NewPatternConstraint 是创建 PatternConstraint 的便捷函数。
// 注意：正则表达式的编译被延迟到第一次 Validate 调用时。
func NewPatternConstraint(pattern string) *PatternConstraint {
	return &PatternConstraint{Pattern: pattern}
}

// Validate 检查给定值是否符合指定的正则表达式模式。
func (c *PatternConstraint) Validate(value interface{}) error {
	c.compile() // 确保正则表达式已编译
	if c.initErr != nil {
		return c.initErr // 如果编译失败，直接返回编译错误
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
	Format *string    // 可选的时间格式字符串（例如 "2006-01-02 15:04:05"）。如果为 nil，则验证时间戳或 time.Time。
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
		case float32:
			t = time.Unix(int64(val), 0)
		case float64:
			t = time.Unix(int64(val), 0)
		case time.Time:
			t = val // 直接使用 time.Time 类型
		default:
			return errors.New("类型错误：期望时间戳（整数）或 time.Time 类型")
		}

		// 简单的有效性检查：如果解析后的时间是零值且不是 Unix 纪元，则可能无效。
		if t.IsZero() && value != 0 && value != int64(0) { // 检查是否是实际的零值，而不是 epoch 0
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
				// 返回带有索引和原始错误的详细错误信息，使用 %w 包装。
				return fmt.Errorf("[%d]:%w", i, err)
			}
		}
	default:
		// 如果不是数组或切片类型，则返回错误。
		return fmt.Errorf("类型错误：期望数组或切片类型，实际为 %T", value)
	}
	return nil // 验证通过
}

type FieldConstraint struct {
	Required  bool
	Validator Validator
}

// StructConstraint 结构体字段验证约束
type StructConstraint struct {
	Fields map[string]FieldConstraint // 字段路径到验证器的映射
}

// NewStructConstraint 创建结构体验证器
func NewStructConstraint(fields map[string]FieldConstraint) *StructConstraint {
	return &StructConstraint{Fields: fields}
}

// Validate 检查给定值是否为 map[string]interface{} 并根据定义的字段验证器进行验证。
func (c *StructConstraint) Validate(value interface{}) error {
	// 结构体的值应为 map[string]interface{}
	mapValue, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("type error: expected map[string]interface{} for struct validation, got %T", value)
	}

	for fieldName, constraint := range c.Fields {
		fieldValue, exists := mapValue[fieldName]

		// 如果字段是必填的，但值不存在，则报错
		if constraint.Required && !exists {
			return fmt.Errorf("missing required field: %s", fieldName)
		}

		// 如果字段存在，或者是可选但提供了值，则验证
		if exists {
			if err := constraint.Validator.Validate(fieldValue); err != nil {
				return fmt.Errorf("field '%s' validation failed: %w", fieldName, err)
			}
		}
	}

	return nil
}

// TypeConstraint 类型约束验证器，用于检查值的实际类型是否符合期望。
type TypeConstraint struct {
	ExpectedType reflect.Type // 期望的 Go 类型
	StrictMode   bool
}

// NewTypeConstraint 是创建 TypeConstraint 的便捷函数。
func NewTypeConstraint(t reflect.Type) *TypeConstraint {
	return &TypeConstraint{ExpectedType: t, StrictMode: false}
}

// NewTypeConstraintWithMode 创建带模式选项的类型约束验证器
func NewTypeConstraintWithMode(t reflect.Type, strict bool) *TypeConstraint {
	return &TypeConstraint{
		ExpectedType: t,
		StrictMode:   strict,
	}
}

// Validate 检查给定值的类型是否可以赋值给期望的类型。
func (c *TypeConstraint) Validate(value interface{}) error {
	if value == nil {
		// 空值直接通过，由上层验证器处理空值逻辑
		return nil
	}

	actual := reflect.TypeOf(value)

	// 1. 严格模式：直接检查类型兼容性
	if c.StrictMode {
		if !actual.AssignableTo(c.ExpectedType) {
			return fmt.Errorf("类型错误：期望类型 %v，实际类型 %v", c.ExpectedType, actual)
		}
		return nil
	}

	// 2. 非严格模式：允许基本类型之间的转换
	expectedKind := c.ExpectedType.Kind()
	actualKind := actual.Kind()

	// 检查类型兼容性（允许转换）
	switch {
	case actual.AssignableTo(c.ExpectedType):
		// 类型完全兼容
		return nil

	case isNumericKind(expectedKind) && isNumericKind(actualKind):
		// 数字类型兼容：int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64
		return nil

	case expectedKind == reflect.String && actualKind == reflect.SliceOf(reflect.TypeOf(byte(0))).Kind():
		// 允许 []byte 转 string
		return nil

	case expectedKind == reflect.SliceOf(reflect.TypeOf(byte(0))).Kind() && actualKind == reflect.String:
		// 允许 string 转 []byte
		return nil

	case expectedKind == reflect.Interface:
		// 期望接口类型，检查是否实现了接口
		if actual.Implements(c.ExpectedType) {
			return nil
		}
	}

	// 所有其他情况都不兼容
	return fmt.Errorf("类型错误：期望类型 %v 或兼容类型，实际类型 %v", c.ExpectedType, actual)
}

// isNumericKind 检查是否为数字类型
func isNumericKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}
