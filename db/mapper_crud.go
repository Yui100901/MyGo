package db

import "gorm.io/gorm"

//
// @Author yfy2001
// @Date 2025/8/21 15 50
//

// Create 创建记录
func (m *Mapper[T]) Create(record *T) *Result[*T] {
	result := m.db.Create(record)
	if result.Error != nil {
		return Fail[*T](result.Error)
	}
	return Ok(record, result.RowsAffected)
}

// CreateBatch 批量创建记录
func (m *Mapper[T]) CreateBatch(records []*T) *Result[[]*T] {
	result := m.db.Create(records)
	if result.Error != nil {
		return Fail[[]*T](result.Error)
	}
	return Ok(records, result.RowsAffected)
}

// FindByID 根据主键查找记录
func (m *Mapper[T]) FindByID(id any) *Result[*T] {
	var record T
	result := m.db.First(&record, id)
	if result.Error != nil {
		return Fail[*T](result.Error)
	}
	return Ok(&record, result.RowsAffected)
}

// FindOne 根据条件查找单条记录
func (m *Mapper[T]) FindOne(conditions map[string]interface{}) *Result[*T] {
	var record T
	result := m.db.Where(conditions).First(&record)
	if result.Error != nil {
		return Fail[*T](result.Error)
	}
	return Ok(&record, result.RowsAffected)
}

// FindAll 根据条件查找所有记录
func (m *Mapper[T]) FindAll(conditions map[string]interface{}) *Result[[]*T] {
	var records []*T
	result := m.db.Where(conditions).Find(&records)
	if result.Error != nil {
		return Fail[[]*T](result.Error)
	}
	return Ok(records, result.RowsAffected)
}

// Update 更新记录
func (m *Mapper[T]) Update(record *T) *Result[*T] {
	result := m.db.Save(record)
	if result.Error != nil {
		return Fail[*T](result.Error)
	}
	return Ok(record, result.RowsAffected)
}

// UpdateSelective 选择性更新（只更新非零值字段）
func (m *Mapper[T]) UpdateSelective(record *T) *Result[*T] {
	result := m.db.Model(record).Updates(record)
	if result.Error != nil {
		return Fail[*T](result.Error)
	}
	return Ok(record, result.RowsAffected)
}

// UpdateWhere 根据条件更新字段
func (m *Mapper[T]) UpdateWhere(conditions map[string]interface{}, updates map[string]interface{}) *Result[int64] {
	result := m.db.Model(new(T)).Where(conditions).Updates(updates)
	if result.Error != nil {
		return Fail[int64](result.Error)
	}
	return Ok(result.RowsAffected, result.RowsAffected)
}

// DeleteById 删除记录
func (m *Mapper[T]) DeleteById(record *T) *Result[int64] {
	result := m.db.Delete(record)
	if result.Error != nil {
		return Fail[int64](result.Error)
	}
	return Ok(result.RowsAffected, result.RowsAffected)
}

// DeleteWhere 根据条件删除记录
func (m *Mapper[T]) DeleteWhere(conditions map[string]interface{}) *Result[int64] {
	result := m.db.Where(conditions).Delete(new(T))
	if result.Error != nil {
		return Fail[int64](result.Error)
	}
	return Ok(result.RowsAffected, result.RowsAffected)
}

// PaginateQuery 分页查询（支持自定义条件和排序）
func (m *Mapper[T]) PaginateQuery(conditions map[string]interface{}, order string, current int, pageSize int) *Result[*Page[T]] {
	// 获取总数
	var total int64
	countResult := m.db.Model(new(T)).Where(conditions).Count(&total)
	if countResult.Error != nil {
		return Fail[*Page[T]](countResult.Error)
	}

	// 获取分页数据
	var records []*T
	query := m.db.Where(conditions)
	if order != "" {
		query = query.Order(order)
	}

	offset := (current - 1) * pageSize
	result := query.Offset(offset).Limit(pageSize).Find(&records)
	if result.Error != nil {
		return Fail[*Page[T]](result.Error)
	}

	// 构建分页结果
	page := Paginate(records, total, current, pageSize)
	return Ok(page, total)
}

// FirstOrCreate 查找第一条记录，如果不存在则创建
func (m *Mapper[T]) FirstOrCreate(conditions map[string]interface{}, record *T) *Result[*T] {
	result := m.db.Where(conditions).FirstOrCreate(record)
	if result.Error != nil {
		return Fail[*T](result.Error)
	}
	return Ok(record, result.RowsAffected)
}

// WithPreload 预加载关联数据
func (m *Mapper[T]) WithPreload(preloads ...string) *Mapper[T] {
	db := m.db
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	return &Mapper[T]{db: db}
}

// WithScope 添加自定义查询条件
func (m *Mapper[T]) WithScope(fn func(*gorm.DB) *gorm.DB) *Mapper[T] {
	return &Mapper[T]{db: fn(m.db)}
}
