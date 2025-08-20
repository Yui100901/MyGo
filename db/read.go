package db

import "gorm.io/gorm"

//
// @Author yfy2001
// @Date 2025/8/20 17 14
//

func (m *Mapper[T]) GetList() *Result[[]*T] {
	var list []*T
	result := m.db.Find(&list)
	if result.Error != nil {
		return Fail[[]*T](result.Error)
	}
	return Ok(list, result.RowsAffected)
}

func (m *Mapper[T]) GetByID(id string) *Result[*T] {
	var t T
	result := m.db.First(&t, "id = ?", id)
	if result.Error != nil {
		return Fail[*T](result.Error)
	}
	return Ok(&t, result.RowsAffected)
}

func (m *Mapper[T]) GetByCondition(conditions map[string]interface{}) *Result[[]*T] {
	var list []*T
	result := m.db.Where(conditions).Find(&list)
	if result.Error != nil {
		return Fail[[]*T](result.Error)
	}
	return Ok(list, result.RowsAffected)
}

func (m *Mapper[T]) GetSortedList(orderBy string) *Result[[]*T] {
	var list []*T
	result := m.db.Order(orderBy).Find(&list)
	if result.Error != nil {
		return Fail[[]*T](result.Error)
	}
	return Ok(list, result.RowsAffected)
}

func (m *Mapper[T]) GetOneByCondition(conditions map[string]interface{}) *Result[*T] {
	var t T
	result := m.db.Where(conditions).First(&t)
	if result.Error != nil {
		return Fail[*T](result.Error)
	}
	return Ok(&t, result.RowsAffected)
}

func (m *Mapper[T]) GetByQuery(query func(*gorm.DB) *gorm.DB) *Result[[]*T] {
	var list []*T
	result := query(m.db).Find(&list)
	if result.Error != nil {
		return Fail[[]*T](result.Error)
	}
	return Ok(list, result.RowsAffected)
}

func (m *Mapper[T]) GetPagedByConditionPage(conditions map[string]interface{}, page, pageSize int, orderBy string) *Result[*Page[T]] {
	var list []*T
	var total int64
	offset := (page - 1) * pageSize

	query := m.db.Model(new(T)).Where(conditions)
	query.Count(&total)

	result := query.Order(orderBy).Limit(pageSize).Offset(offset).Find(&list)
	if result.Error != nil {
		return Fail[*Page[T]](result.Error)
	}
	return Ok(Paginate(list, total, page, pageSize), total)
}

func (m *Mapper[T]) GetPaginatedListPage(page, pageSize int, orderBy string) *Result[*Page[T]] {
	var list []*T
	var total int64
	offset := (page - 1) * pageSize

	query := m.db.Model(new(T))
	query.Count(&total)

	result := query.Order(orderBy).Limit(pageSize).Offset(offset).Find(&list)
	if result.Error != nil {
		return Fail[*Page[T]](result.Error)
	}
	return Ok(Paginate(list, total, page, pageSize), total)
}
