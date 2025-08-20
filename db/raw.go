package db

//
// @Author yfy2001
// @Date 2025/8/18 22 03
//

// ================================= 原生SQL操作 =================================

// Raw 执行原生SQL查询
func (m *Mapper[T]) Raw(sql string, args ...interface{}) *Result[[]*T] {
	var list []*T
	res := m.db.Raw(sql, args...).Scan(&list)
	if res.Error != nil {
		return Fail[[]*T](res.Error)
	}
	return Ok(list, res.RowsAffected)
}

// Exec 执行原生SQL命令
func (m *Mapper[T]) Exec(sql string, args ...interface{}) *Result[any] {
	res := m.db.Exec(sql, args...)
	if res.Error != nil {
		return Fail[any](res.Error)
	}
	return Ok[any](nil, res.RowsAffected)
}
