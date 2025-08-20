package db

import "fmt"

//
// @Author yfy2001
// @Date 2025/8/20 17 15
//

func (m *Mapper[T]) DeleteByID(id string) *Result[any] {
	var t T
	result := m.db.Delete(&t, "id = ?", id)
	if result.Error != nil {
		return Fail[any](result.Error)
	}
	return Ok(any(nil), result.RowsAffected)
}

func (m *Mapper[T]) BatchDelete(conditions map[string]interface{}) *Result[any] {
	var t T
	result := m.db.Where(conditions).Delete(&t)
	if result.Error != nil {
		return Fail[any](result.Error)
	}
	return Ok(any(nil), result.RowsAffected)
}

func (m *Mapper[T]) BatchDeleteByIdList(idList []any) *Result[any] {
	if len(idList) == 0 {
		return Fail[any](fmt.Errorf("idList 不能为空"))
	}
	var t T
	result := m.db.Where("id IN ?", idList).Delete(&t)
	if result.Error != nil {
		return Fail[any](result.Error)
	}
	return Ok(any(nil), result.RowsAffected)
}

func (m *Mapper[T]) RestoreByID(id string) *Result[any] {
	var t T
	result := m.db.Unscoped().Model(&t).Where("id = ?", id).Update("deleted_at", nil)
	if result.Error != nil {
		return Fail[any](result.Error)
	}
	return Ok(any(nil), result.RowsAffected)
}
