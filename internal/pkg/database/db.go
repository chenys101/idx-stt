package database

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"stt/internal/model"
)

var DB *gorm.DB

func Connect(dsn string) error {

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: false, // 启用外键
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %v", err)
	}
	// 设置连接池（SQLite专用优化）
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1) // SQLite写操作单连接限制
	sqlDB.SetMaxIdleConns(5)

	// Auto migrate models
	if err := db.AutoMigrate(&model.User{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&model.StockMonitor{}); err != nil {
		return err
	}
	DB = db
	return nil
}
