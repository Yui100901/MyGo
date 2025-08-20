package db

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"log"
	"os"
	"sync"
	"time"
)

type Model interface {
	TableName() string
}

type DatabaseType string

const (
	SQLITE    DatabaseType = "sqlite"
	MYSQL     DatabaseType = "mysql"
	SQLSERVER DatabaseType = "mssql"
	POSTGRES  DatabaseType = "postgres"
)

var (
	dbInstance *gorm.DB
	once       sync.Once
	logger     = log.New(os.Stdout, "[DB] ", log.LstdFlags|log.Lshortfile)
)

// 驱动注册表
var drivers = map[DatabaseType]func(dsn string) gorm.Dialector{
	SQLITE:    func(dsn string) gorm.Dialector { return sqlite.Open(dsn) },
	MYSQL:     func(dsn string) gorm.Dialector { return mysql.Open(dsn) },
	SQLSERVER: func(dsn string) gorm.Dialector { return sqlserver.Open(dsn) },
	POSTGRES:  func(dsn string) gorm.Dialector { return postgres.Open(dsn) },
}

// InitDB 初始化数据库（单例）
func InitDB(dbType DatabaseType, dsn string) *gorm.DB {
	once.Do(func() {
		dialect, ok := drivers[dbType]
		if !ok {
			panic(fmt.Sprintf("unsupported database type: %s", dbType))
		}

		db, err := connectDB(dialect(dsn), string(dbType))
		if err != nil {
			panic(fmt.Sprintf("failed to initialize database: %v", err))
		}
		dbInstance = db
	})
	return dbInstance
}

// 统一连接函数 + 连接池配置
func connectDB(dialector gorm.Dialector, name string) (*gorm.DB, error) {
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		logger.Printf("Failed to connect %s!", name)
		return nil, err
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	logger.Printf("Connected to %s!", name)
	return db, nil
}
