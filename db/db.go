package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Yui100901/MyGo/concurrency"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type DatabaseType string

const (
	SQLITE    DatabaseType = "sqlite"
	MYSQL     DatabaseType = "mysql"
	SQLSERVER DatabaseType = "mssql"
	POSTGRES  DatabaseType = "postgres"
)

var (
	logger = log.New(os.Stdout, "[DB] ", log.LstdFlags|log.Lshortfile)
	// 全局线程安全 Map
	dbMap = concurrency.NewSafeMap[string, *gorm.DB](32)
)

// 驱动注册表
var drivers = map[DatabaseType]func(dsn string) gorm.Dialector{
	SQLITE:    sqlite.Open,
	MYSQL:     mysql.Open,
	SQLSERVER: sqlserver.Open,
	POSTGRES:  postgres.Open,
}

// InitDB 初始化数据库（单例）
func InitDB(name string, dbType DatabaseType, dsn string) *gorm.DB {
	if db, ok := dbMap.Get(name); ok {
		return db
	}
	dialect, ok := drivers[dbType]
	if !ok {
		panic(fmt.Sprintf("unsupported database type: %s", dbType))
	}
	db, err := connectDB(dialect(dsn), string(dbType))
	if err != nil {
		panic(fmt.Sprintf("failed to initialize database: %v", err))
	}
	dbMap.Set(name, db)
	return db
}

// RegisterDB 直接注册已有 *gorm.DB（不走 InitDB 流程）
func RegisterDB(name string, db *gorm.DB) error {
	if _, ok := dbMap.Get(name); ok {
		return fmt.Errorf("DB %s already exists", name)
	}
	dbMap.Set(name, db)
	logger.Printf("Registered existing DB %s", name)
	return nil
}

// GetDB 获取数据库连接（存在才返回，否则报错）
func GetDB(name string) (*gorm.DB, error) {
	db, ok := dbMap.Get(name)
	if !ok {
		return nil, fmt.Errorf("no DB found for name: %s", name)
	}
	return db, nil
}

// MustGetDB 获取数据库连接（不存在直接 panic，适合初始化期）
func MustGetDB(name string) *gorm.DB {
	db := dbMap.MustGet(name)
	return db
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
