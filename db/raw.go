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

// RawQuery - 执行原生 SQL 查询
func (m *Mapper[T]) RawQuery(sql string, values ...any) *Result[[]map[string]any] {
	rows, err := m.db.Raw(sql, values...).Rows()
	if err != nil {
		return Fail[[]map[string]any](err)
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	resultList := make([]map[string]any, 0)

	for rows.Next() {
		columns := make([]any, len(cols))
		columnPointers := make([]any, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			return Fail[[]map[string]any](err)
		}
		mapped := make(map[string]any)
		for i, colName := range cols {
			val := columnPointers[i].(*any)
			mapped[colName] = *val
		}
		resultList = append(resultList, mapped)
	}
	return Ok(resultList, int64(len(resultList)))
}
