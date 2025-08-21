package db

import "gorm.io/gorm"

//
// @Author yfy2001
// @Date 2025/8/21 15 52
//

// Where 添加 WHERE 条件
func (m *Mapper[T]) Where(query interface{}, args ...interface{}) *Mapper[T] {
	return &Mapper[T]{db: m.db.Where(query, args...)}
}

// Or 添加 OR 条件
func (m *Mapper[T]) Or(query interface{}, args ...interface{}) *Mapper[T] {
	return &Mapper[T]{db: m.db.Or(query, args...)}
}

// Not 添加 NOT 条件
func (m *Mapper[T]) Not(query interface{}, args ...interface{}) *Mapper[T] {
	return &Mapper[T]{db: m.db.Not(query, args...)}
}

// Order 添加排序条件
func (m *Mapper[T]) Order(value interface{}) *Mapper[T] {
	return &Mapper[T]{db: m.db.Order(value)}
}

// Limit 限制返回记录数
func (m *Mapper[T]) Limit(limit int) *Mapper[T] {
	return &Mapper[T]{db: m.db.Limit(limit)}
}

// Offset 设置偏移量
func (m *Mapper[T]) Offset(offset int) *Mapper[T] {
	return &Mapper[T]{db: m.db.Offset(offset)}
}

// Select 指定要查询的字段
func (m *Mapper[T]) Select(query interface{}, args ...interface{}) *Mapper[T] {
	return &Mapper[T]{db: m.db.Select(query, args...)}
}

// Omit 指定要排除的字段
func (m *Mapper[T]) Omit(columns ...string) *Mapper[T] {
	return &Mapper[T]{db: m.db.Omit(columns...)}
}

// Joins 添加 JOIN 子句
func (m *Mapper[T]) Joins(query string, args ...interface{}) *Mapper[T] {
	return &Mapper[T]{db: m.db.Joins(query, args...)}
}

// Group 添加 GROUP BY 子句
func (m *Mapper[T]) Group(name string) *Mapper[T] {
	return &Mapper[T]{db: m.db.Group(name)}
}

// Having 添加 HAVING 子句
func (m *Mapper[T]) Having(query interface{}, args ...interface{}) *Mapper[T] {
	return &Mapper[T]{db: m.db.Having(query, args...)}
}

// Distinct 添加 DISTINCT 子句
func (m *Mapper[T]) Distinct(args ...interface{}) *Mapper[T] {
	return &Mapper[T]{db: m.db.Distinct(args...)}
}

// Preload 预加载关联数据
func (m *Mapper[T]) Preload(query string, args ...interface{}) *Mapper[T] {
	return &Mapper[T]{db: m.db.Preload(query, args...)}
}

// Scopes 应用作用域
func (m *Mapper[T]) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *Mapper[T] {
	return &Mapper[T]{db: m.db.Scopes(funcs...)}
}

// Unscoped 包含软删除记录
func (m *Mapper[T]) Unscoped() *Mapper[T] {
	return &Mapper[T]{db: m.db.Unscoped()}
}

// Debug 开启调试模式
func (m *Mapper[T]) Debug() *Mapper[T] {
	return &Mapper[T]{db: m.db.Debug()}
}

// 终结方法 - 执行查询并返回结果

// First 获取第一条记录
func (m *Mapper[T]) First() *Result[*T] {
	var record T
	result := m.db.First(&record)
	if result.Error != nil {
		return Fail[*T](result.Error)
	}
	return Ok(&record, result.RowsAffected)
}

// Last 获取最后一条记录
func (m *Mapper[T]) Last() *Result[*T] {
	var record T
	result := m.db.Last(&record)
	if result.Error != nil {
		return Fail[*T](result.Error)
	}
	return Ok(&record, result.RowsAffected)
}

// Find 获取所有匹配记录
func (m *Mapper[T]) Find() *Result[[]*T] {
	var records []*T
	result := m.db.Find(&records)
	if result.Error != nil {
		return Fail[[]*T](result.Error)
	}
	return Ok(records, result.RowsAffected)
}

// Count 统计记录数
func (m *Mapper[T]) Count() *Result[int64] {
	var count int64
	result := m.db.Count(&count)
	if result.Error != nil {
		return Fail[int64](result.Error)
	}
	return Ok(count, count)
}

// Pluck 查询单列值
func (m *Mapper[T]) Pluck(column string) *Result[[]interface{}] {
	var values []interface{}
	result := m.db.Pluck(column, &values)
	if result.Error != nil {
		return Fail[[]interface{}](result.Error)
	}
	return Ok(values, result.RowsAffected)
}

// UpdateColumns 更新指定字段
func (m *Mapper[T]) UpdateColumns(values interface{}) *Result[int64] {
	result := m.db.UpdateColumns(values)
	if result.Error != nil {
		return Fail[int64](result.Error)
	}
	return Ok(result.RowsAffected, result.RowsAffected)
}

// Delete 删除匹配记录
func (m *Mapper[T]) Delete() *Result[int64] {
	result := m.db.Delete(new(T))
	if result.Error != nil {
		return Fail[int64](result.Error)
	}
	return Ok(result.RowsAffected, result.RowsAffected)
}

// Paginate 分页查询（链式版本）
func (m *Mapper[T]) Paginate(current int, pageSize int) *Result[*Page[T]] {
	// 获取总数
	var total int64
	countResult := m.db.Model(new(T)).Count(&total)
	if countResult.Error != nil {
		return Fail[*Page[T]](countResult.Error)
	}

	// 获取分页数据
	var records []*T
	offset := (current - 1) * pageSize
	result := m.db.Offset(offset).Limit(pageSize).Find(&records)
	if result.Error != nil {
		return Fail[*Page[T]](result.Error)
	}

	// 构建分页结果
	page := Paginate(records, total, current, pageSize)
	return Ok(page, total)
}
