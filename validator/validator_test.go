package validator

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 测试 RangeConstraint
func TestRangeConstraint(t *testing.T) {
	t.Run("int类型验证", func(t *testing.T) {
		min := 10
		max := 20
		c := &RangeConstraint[int]{Min: &min, Max: &max}

		assert.NoError(t, c.Validate(15))
		assert.ErrorContains(t, c.Validate(5), "小于最小值")
		assert.ErrorContains(t, c.Validate(25), "大于最大值")
		assert.ErrorContains(t, c.Validate("string"), "类型错误")
	})

	t.Run("float类型验证", func(t *testing.T) {
		min := 1.5
		c := &RangeConstraint[float64]{Min: &min}

		assert.NoError(t, c.Validate(2.0))
		assert.ErrorContains(t, c.Validate(1.0), "小于最小值")
	})

	t.Run("字符串范围", func(t *testing.T) {
		min := "apple"
		max := "cherry"
		c := &RangeConstraint[string]{Min: &min, Max: &max}

		assert.NoError(t, c.Validate("banana"))
		assert.ErrorContains(t, c.Validate("date"), "大于最大值")
	})
}

// 测试 LengthConstraint
func TestLengthConstraint(t *testing.T) {
	t.Run("字符串长度", func(t *testing.T) {
		min := 3
		max := 5
		c := &LengthConstraint{Min: &min, Max: &max}

		assert.NoError(t, c.Validate("abc"))
		assert.NoError(t, c.Validate("abcde"))
		assert.ErrorContains(t, c.Validate("ab"), "小于最小值")
		assert.ErrorContains(t, c.Validate("abcdef"), "大于最大值")
	})

	t.Run("切片长度", func(t *testing.T) {
		min := 2
		c := &LengthConstraint{Min: &min}

		assert.NoError(t, c.Validate([]int{1, 2}))
		assert.ErrorContains(t, c.Validate([]string{"a"}), "小于最小值")
	})

	t.Run("指针类型处理", func(t *testing.T) {
		min := 1
		c := &LengthConstraint{Min: &min}

		str := "hello"
		assert.NoError(t, c.Validate(&str))

		var nilPtr *string
		assert.ErrorContains(t, c.Validate(nilPtr), "小于最小值")
	})

	t.Run("不支持的类型", func(t *testing.T) {
		c := &LengthConstraint{}
		err := c.Validate(42)
		assert.ErrorContains(t, err, "不支持长度验证的类型")
	})
}

// 测试 EnumConstraint
func TestEnumConstraint(t *testing.T) {
	t.Run("字符串枚举", func(t *testing.T) {
		c := NewEnumConstraint("red", "green", "blue")

		assert.NoError(t, c.Validate("green"))
		assert.ErrorContains(t, c.Validate("yellow"), "不在枚举范围内")
	})

	t.Run("整数枚举", func(t *testing.T) {
		c := NewEnumConstraint(1, 3, 5)

		assert.NoError(t, c.Validate(3))
		assert.ErrorContains(t, c.Validate(2), "不在枚举范围内")
	})

	t.Run("空枚举集合", func(t *testing.T) {
		c := NewEnumConstraint[string]()
		err := c.Validate("any")
		assert.ErrorContains(t, err, "枚举约束配置无效")
	})
}

// 测试 RequiredConstraint
func TestRequiredConstraint(t *testing.T) {
	c := &RequiredConstraint{}

	t.Run("非空值", func(t *testing.T) {
		assert.NoError(t, c.Validate("text"))
		assert.NoError(t, c.Validate([]int{1, 2}))
		assert.NoError(t, c.Validate(42))
	})

	t.Run("空值", func(t *testing.T) {
		assert.Error(t, c.Validate(nil))
		assert.Error(t, c.Validate(""))
		assert.Error(t, c.Validate([]int{}))

		var ptr *int
		assert.Error(t, c.Validate(ptr))
	})
}

// 测试 PatternConstraint
func TestPatternConstraint(t *testing.T) {
	t.Run("有效模式", func(t *testing.T) {
		c := NewPatternConstraint(`^[a-z]+$`)

		assert.NoError(t, c.Validate("abc"))
		assert.ErrorContains(t, c.Validate("123"), "不符合格式要求")
	})

	t.Run("字节切片", func(t *testing.T) {
		c := NewPatternConstraint(`^\d+$`)
		assert.NoError(t, c.Validate([]byte("123")))
	})

	t.Run("无效模式", func(t *testing.T) {
		c := NewPatternConstraint(`[invalid`)
		err := c.Validate("any")
		assert.Error(t, err)
	})

	t.Run("空模式", func(t *testing.T) {
		c := NewPatternConstraint("")
		err := c.Validate("any")
		assert.ErrorContains(t, err, "正则表达式不能为空")
	})
}

// 测试 TimeConstraint
func TestTimeConstraint(t *testing.T) {
	now := time.Now()
	format := "2006-01-02"

	t.Run("时间戳验证", func(t *testing.T) {
		min := now.Add(-time.Hour)
		max := now.Add(time.Hour)
		c := &TimeConstraint{Min: &min, Max: &max}

		assert.NoError(t, c.Validate(now.Unix()))
		assert.ErrorContains(t, c.Validate(now.Add(-2*time.Hour).Unix()), "早于最小值")
	})

	t.Run("格式验证", func(t *testing.T) {
		c := &TimeConstraint{Format: &format}
		assert.NoError(t, c.Validate("2023-01-01"))
		assert.ErrorContains(t, c.Validate("01-01-2023"), "时间格式不符合要求")
	})

	t.Run("无效类型", func(t *testing.T) {
		c := &TimeConstraint{}
		assert.ErrorContains(t, c.Validate("string"), "类型错误")
	})
}

// 测试 ArrayConstraint
func TestArrayConstraint(t *testing.T) {
	t.Run("元素验证", func(t *testing.T) {
		min := 5
		itemValidator := &RangeConstraint[int]{Min: &min}
		c := &ArrayConstraint{Item: itemValidator}

		assert.NoError(t, c.Validate([]int{6, 7, 8}))
		err := c.Validate([]int{3, 6, 9})
		assert.ErrorContains(t, err, "元素 [0] 验证失败")
	})

	t.Run("空验证器", func(t *testing.T) {
		c := &ArrayConstraint{}
		err := c.Validate([]int{1, 2, 3})
		assert.ErrorContains(t, err, "数组元素验证器不能为空")
	})

	t.Run("非数组类型", func(t *testing.T) {
		min := 0
		itemValidator := &RangeConstraint[int]{Min: &min}
		c := &ArrayConstraint{Item: itemValidator}

		err := c.Validate("not an array")
		assert.ErrorContains(t, err, "期望数组或切片类型")
	})
}

// 测试 TypeConstraint
func TestTypeConstraint(t *testing.T) {
	t.Run("类型匹配", func(t *testing.T) {
		c := NewTypeConstraint(reflect.TypeOf(0))
		assert.NoError(t, c.Validate(42))
		assert.ErrorContains(t, c.Validate("string"), "类型错误")
	})

}

// 测试 ConditionalValidator
func TestConditionalValidator(t *testing.T) {
	t.Run("条件满足", func(t *testing.T) {
		condition := func(v interface{}) bool {
			return v.(int) > 0
		}

		positiveValidator := NewFuncValidator(func(v interface{}) error {
			if v.(int) > 100 {
				return errors.New("不能超过100")
			}
			return nil
		}, "正数检查")

		c := NewConditionalValidator(condition, positiveValidator, "条件验证")

		assert.NoError(t, c.Validate(50))
		assert.ErrorContains(t, c.Validate(150), "不能超过100")
	})

	t.Run("条件不满足", func(t *testing.T) {
		condition := func(v interface{}) bool {
			return v.(int) < 0
		}

		validator := NewFuncValidator(func(v interface{}) error {
			return errors.New("永远不会执行")
		}, "负值检查")

		c := NewConditionalValidator(condition, validator, "条件验证")
		assert.NoError(t, c.Validate(10)) // 条件不满足，跳过验证
	})
}

// 测试 FuncValidator
func TestFuncValidator(t *testing.T) {
	t.Run("自定义验证", func(t *testing.T) {
		v := NewFuncValidator(func(value interface{}) error {
			s, ok := value.(string)
			if !ok {
				return errors.New("需要字符串类型")
			}
			if len(s) < 5 {
				return errors.New("太短")
			}
			return nil
		}, "自定义检查")

		assert.NoError(t, v.Validate("valid string"))
		err := v.Validate("shor")
		assert.ErrorContains(t, err, "太短")
		assert.ErrorContains(t, err, "自定义检查")
	})

	t.Run("未定义函数", func(t *testing.T) {
		v := &FuncValidator{}
		err := v.Validate("anything")
		assert.ErrorContains(t, err, "验证函数未定义")
	})
}

// 测试全局Validate函数
func TestGlobalValidate(t *testing.T) {
	t.Run("多个验证器", func(t *testing.T) {
		min := 10
		max := 20
		rangeValidator := &RangeConstraint[int]{Min: &min, Max: &max}
		evenValidator := NewFuncValidator(func(v interface{}) error {
			if v.(int)%2 != 0 {
				return errors.New("必须为偶数")
			}
			return nil
		}, "偶数检查")

		err := Validate(15, rangeValidator, evenValidator)
		assert.ErrorContains(t, err, "必须为偶数")
	})

	t.Run("短路行为", func(t *testing.T) {
		v1 := NewFuncValidator(func(v interface{}) error {
			return errors.New("first error")
		}, "step1")

		v2 := NewFuncValidator(func(v interface{}) error {
			t.Error("不应该执行第二个验证器")
			return nil
		}, "step2")

		err := Validate(nil, v1, v2)
		assert.Error(t, err)
	})
}

// 测试组合验证器链
func TestValidatorChain(t *testing.T) {
	// 创建一组验证器
	required := &RequiredConstraint{}
	minLength := &LengthConstraint{Min: ptr(5)}
	pattern := NewPatternConstraint(`^[A-Z][a-z]+$`)

	// 组合验证器
	nameValidator := NewCompositeValidator(
		required,
		minLength,
		pattern,
	)

	t.Run("有效名称", func(t *testing.T) {
		assert.NoError(t, nameValidator.Validate("Alexander"))
	})

	t.Run("无效名称", func(t *testing.T) {
		tests := []struct {
			value string
			error string
		}{
			{"", "不能为空"},
			{"Alex", "小于最小值"},
			{"alexander", "不符合格式要求"},
			{"12345", "不符合格式要求"},
		}

		for _, tc := range tests {
			err := nameValidator.Validate(tc.value)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.error)
		}
	})
}

// 辅助函数：创建指针
func ptr[T any](v T) *T {
	return &v
}
