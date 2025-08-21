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

// GetOrInitDB 初始化数据库（单例）
func GetOrInitDB(name string, dbType DatabaseType, dsn string) (*gorm.DB, error) {
	if db, ok := dbMap.Get(name); ok {
		return db, nil
	}
	logger.Printf("%s not found,try connect:type:%s,dsn:%s", name, dbType, dsn)
	dialect, ok := drivers[dbType]
	if !ok {
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
	db, err := connectDB(dialect(dsn), string(dbType))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	dbMap.Set(name, db)
	return db, nil
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
