package validator

//
// @Author yfy2001
// @Date 2025/7/10 11 19
//

// Validator 数据验证接口
type Validator interface {
	Validate(value any) error
}
