package db

import (
	"gorm.io/gorm"
)

//
// @Author yfy2001
// @Date 2025/7/21 15 32
//

type Model interface {
	TableName() string
}

// Mapper 通用数据访问层
type Mapper[T Model] struct {
	db    *gorm.DB
	model T
}

func NewMapper[T Model](db *gorm.DB) *Mapper[T] {
	return &Mapper[T]{
		db: db,
	}
}

// GetDB 获取底层数据库连接
func (m *Mapper[T]) GetDB() *gorm.DB {
	return m.db
}

// Transaction 执行事务
func (m *Mapper[T]) Transaction(fn func(tx *Mapper[T]) error) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		mapper := NewMapper[T](tx)
		return fn(mapper)
	})
}
