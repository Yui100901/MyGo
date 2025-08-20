package db

//
// @Author yfy2001
// @Date 2025/8/20 17 15
//

func (m *Mapper[T]) Update(id string, updates map[string]interface{}) *Result[any] {
	var t T
	result := m.db.Model(&t).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return Fail[any](result.Error)
	}
	return Ok(any(nil), result.RowsAffected)
}

func (m *Mapper[T]) BatchUpdate(ids []any, updates map[string]interface{}) *Result[any] {
	var t T
	result := m.db.Model(&t).Where("id IN ?", ids).Updates(updates)
	if result.Error != nil {
		return Fail[any](result.Error)
	}
	return Ok(any(nil), result.RowsAffected)
}
