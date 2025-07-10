package validator

import (
	"errors"
	"fmt"
)

//
// @Author yfy2001
// @Date 2025/7/10 11 19
//

// Validator 数据验证接口
type Validator interface {
	Validate(value any) error
}

// NewCompositeValidator 是创建 CompositeValidator 的便捷函数。
func NewCompositeValidator(validators ...Validator) *CompositeValidator {
	return &CompositeValidator{validators: validators}
}

// Validate 依次执行所有组合的验证器。如果任何一个验证器失败，则立即返回错误。
func (cv *CompositeValidator) Validate(value interface{}) error {
	for _, v := range cv.validators {
		if err := v.Validate(value); err != nil {
			// 返回的错误包含是哪个验证器失败以及具体的失败信息，使用 %w 包装原始错误。
			return fmt.Errorf("%w", err)
		}
	}
	return nil // 所有验证器都通过
}

// ConditionalValidator 根据条件决定是否应用验证
type ConditionalValidator struct {
	condition   func(interface{}) bool
	validator   Validator
	description string
}

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

func (v *ConditionalValidator) Validate(value interface{}) error {
	if v.condition != nil && v.condition(value) {
		if err := v.validator.Validate(value); err != nil {
			return fmt.Errorf("%s: %w", v.description, err)
		}
	}
	return nil
}

// FuncValidator 允许使用自定义函数进行验证
type FuncValidator struct {
	validateFunc func(interface{}) error
	description  string
}

func NewFuncValidator(fn func(interface{}) error, desc string) *FuncValidator {
	return &FuncValidator{
		validateFunc: fn,
		description:  desc,
	}
}

func (v *FuncValidator) Validate(value interface{}) error {
	if v.validateFunc == nil {
		return errors.New("验证函数未定义")
	}
	if err := v.validateFunc(value); err != nil {
		return fmt.Errorf("%s: %w", v.description, err)
	}
	return nil
}

// Validate 全局验证函数，作为一种便捷的工具函数，直接传入要验证的值和一系列验证器。
func Validate(value interface{}, validators ...Validator) error {
	for _, v := range validators {
		if err := v.Validate(value); err != nil {
			return err // 任何一个验证器失败，立即返回错误
		}
	}
	return nil // 所有验证器都通过
}
