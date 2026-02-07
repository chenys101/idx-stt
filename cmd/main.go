package main

import (
	"fmt"
	"log"
	"stt/internal/config"
	"stt/internal/pkg/database"
	"stt/internal/route"
)

func main() {
	// 加载配置
	cfg, err := config.Load("./config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	if err := database.Connect(cfg.Database.DSN); err != nil {
		log.Fatalf("Database connection error: %v", err)
	}

	// 创建路由
	r := route.SetupRouter()

	// 启动服务
	if err := r.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}
