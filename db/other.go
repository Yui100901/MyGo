package db

//
// @Author yfy2001
// @Date 2025/8/20 17 19
//

// Count - 按条件统计记录数
func (m *Mapper[T]) Count(conditions map[string]interface{}) *Result[int64] {
	var total int64
	result := m.db.Model(new(T)).Where(conditions).Count(&total)
	if result.Error != nil {
		return Fail[int64](result.Error)
	}
	return Ok(total, total)
}

// Exists - 判断是否存在符合条件的记录
func (m *Mapper[T]) Exists(conditions map[string]interface{}) *Result[bool] {
	var count int64
	result := m.db.Model(new(T)).Where(conditions).Count(&count)
	if result.Error != nil {
		return Fail[bool](result.Error)
	}
	return Ok(count > 0, count)
}
