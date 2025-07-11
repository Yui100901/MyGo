package validator

import (
	"errors"
	"fmt"
)

// @Author yfy2001
// @Date 2025/7/10 11:19

// Validator 数据验证接口定义了所有验证器必须实现的方法。
// 所有具体的约束（如 RangeConstraint、LengthConstraint 等）都应实现此接口。
type Validator interface {
	Validate(value interface{}) error
}

// CompositeValidator 组合验证器，允许将多个验证器链式应用于同一个值。
// 如果一个值需要满足多个不同的验证规则（例如，既要满足长度约束，又要满足正则约束），
// 可以使用 CompositeValidator 将这些规则组合起来。
type CompositeValidator struct {
	validators []Validator // 包含的验证器切片
}

// NewCompositeValidator 是创建 CompositeValidator 的便捷函数。
// 它接收可变参数的 Validator 接口，并返回一个新的 CompositeValidator 实例。
func NewCompositeValidator(validators ...Validator) *CompositeValidator {
	return &CompositeValidator{validators: validators}
}

// Add 为组合验证器添加验证器
func (cv *CompositeValidator) Add(validators ...Validator) {
	cv.validators = append(cv.validators, validators...)
}

// Validate 依次执行所有组合的验证器。
// 如果任何一个验证器失败，则立即返回错误，并使用 %w 包装原始错误，以便进行错误链追踪。
// 如果所有验证器都通过，则返回 nil。
func (cv *CompositeValidator) Validate(value interface{}) error {
	for _, v := range cv.validators {
		if err := v.Validate(value); err != nil {
			// 返回的错误包含是哪个验证器失败以及具体的失败信息，使用 %w 包装原始错误。
			// 这允许调用者检查错误链，例如使用 errors.Is 或 errors.As。
			return fmt.Errorf("%w", err)
		}
	}
	return nil // 所有验证器都通过
}

// ConditionalValidator 根据条件决定是否应用验证。
// 它允许您定义一个在执行实际验证器之前检查的条件函数。
// 如果条件为真，则应用内部的 validator；否则，跳过验证。
type ConditionalValidator struct {
	condition   func(interface{}) bool // 决定是否执行验证的条件函数
	validator   Validator              // 在条件为真时执行的实际验证器
	description string                 // 当验证失败时，作为错误信息前缀的描述
}

// NewConditionalValidator 是创建 ConditionalValidator 的便捷函数。
// condition: 一个函数，如果返回 true，则 validator 将被执行。
// validator: 在条件满足时要执行的 Validator。
// desc: 一个描述性字符串，用于在验证失败时提供上下文。
func NewConditionalValidator(
	condition func(interface{}) bool,
	validator Validator,
	desc string,
) *ConditionalValidator {
	return &ConditionalValidator{
		condition:   condition,
		validator:   validator,
		description: desc,
	}
}

// Validate 根据预定义的条件执行验证。
// 如果 condition 为 nil 或者 condition(value) 返回 true，则执行内部 validator 的 Validate 方法。
// 如果内部 validator 失败，则返回一个带有描述和原始错误的包装错误。
func (v *ConditionalValidator) Validate(value interface{}) error {
	if v.condition != nil && v.condition(value) {
		if err := v.validator.Validate(value); err != nil {
			return fmt.Errorf("%s: %w", v.description, err)
		}
	}
	// 如果条件不满足或条件函数为 nil，则认为验证通过（不执行内部验证）
	return nil
}

// FuncValidator 允许使用自定义函数进行验证。
// 这对于需要特殊逻辑的验证场景非常有用，避免为每个自定义验证创建新的结构体。
type FuncValidator struct {
	validateFunc func(interface{}) error // 自定义的验证逻辑函数
	description  string                  // 当验证失败时，作为错误信息前缀的描述
}

// NewFuncValidator 是创建 FuncValidator 的便捷函数。
// fn: 自定义的验证函数，它接收一个值并返回一个错误（如果验证失败）。
// desc: 一个描述性字符串，用于在验证失败时提供上下文。
func NewFuncValidator(fn func(interface{}) error, desc string) *FuncValidator {
	return &FuncValidator{
		validateFunc: fn,
		description:  desc,
	}
}

// Validate 执行自定义的验证函数。
// 如果 validateFunc 为 nil，则返回一个错误。
// 如果 validateFunc 返回非 nil 错误，则返回一个带有描述和原始错误的包装错误。
func (v *FuncValidator) Validate(value interface{}) error {
	if v.validateFunc == nil {
		return errors.New("验证函数未定义")
	}
	if err := v.validateFunc(value); err != nil {
		return fmt.Errorf("%s: %w", v.description, err)
	}
	return nil
}

// Validate 全局验证函数，作为一种便捷的工具函数。
// 它接收一个要验证的值和一系列 Validator 接口。
// 依次执行所有传入的验证器。如果任何一个验证器失败，则立即返回错误。
// 此函数简化了对单个值应用多个验证器的过程，常用于顶层验证。
func Validate(value interface{}, validators ...Validator) error {
	for _, v := range validators {
		if err := v.Validate(value); err != nil {
			return err // 任何一个验证器失败，立即返回错误
		}
	}
	return nil // 所有验证器都通过
}
