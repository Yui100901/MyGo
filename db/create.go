package db

//
// @Author yfy2001
// @Date 2025/8/20 17 13
//

func (m *Mapper[T]) SaveOrUpdate(t *T) *Result[*T] {
	result := m.db.Save(t)
	if result.Error != nil {
		return Fail[*T](result.Error)
	}
	return Ok(t, result.RowsAffected)
}

func (m *Mapper[T]) BatchInsert(records []*T) *Result[[]*T] {
	result := m.db.Create(&records)
	if result.Error != nil {
		return Fail[[]*T](result.Error)
	}
	return Ok(records, result.RowsAffected)
}
