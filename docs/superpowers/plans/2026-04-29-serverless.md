# Serverless 改造评估 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不考虑/不保证用户模块（`/api/v1/users`）的前提下，评估并（可选）最小改造当前 Gin + SQLite 服务，使其能在主流 Serverless（以 AWS Lambda + API Gateway 为参考）模式下运行，同时保留原有本地长驻进程运行方式；重点确保股票相关接口（`/api/v1/std/*`）可用。

**Architecture:** 将“HTTP 路由构建”与“进程启动/事件入口”解耦：保留现有 `cmd/main.go` 的长驻模式，同时新增一个 Serverless 入口（Lambda handler），复用同一套 `route.SetupRouter()`。数据库层从“本地 data.db”切换为“可配置的外部数据库/持久化方案”，并避免在每次 cold start 做破坏性迁移。

**Tech Stack:** Go 1.24、Gin、GORM、Viper；（新增）AWS Lambda 适配库 `github.com/awslabs/aws-lambda-go-api-proxy/gin` 与 `github.com/aws/aws-lambda-go/lambda`（仅当选择 AWS Lambda 路线时）。

---

## 0. 评估结论（先给答案）

### 0.1 可以改造成 Serverless 吗？

可以，但需要解决两个关键问题：

1. **运行模型差异**
   - 现状：`r.Run(":port")` 需要常驻监听端口（[main.go](file:///workspace/cmd/main.go#L23-L29)）
   - Serverless：平台把请求以事件方式交给函数，不允许你自行长期 listen；要用“事件 → http.Handler”适配层
2. **SQLite 本地文件不适合函数持久化写入**
   - 现状：默认 `dsn: "data.db"`，并且 Dockerfile 会把 `data.db` 拷入镜像（[config.yaml](file:///workspace/config.yaml#L1-L5)，[Dockerfile](file:///workspace/Dockerfile#L23-L25)）
   - Serverless：文件系统通常是只读或短暂的（即便可写也会丢失），而股票监控接口会写入 `stock_monitors` 表（[StockMonitor](file:///workspace/internal/model/stock.go#L15-L20)）

### 0.2 推荐路线（从易到难）

**路线 A（推荐）：AWS Lambda + API Gateway + 外部数据库**

- 适合：需要真正的函数式伸缩 + 按调用计费
- 代价：需要引入 Lambda 适配与外部 DB（MySQL/Postgres 等），并处理连接复用与迁移策略

**路线 B：Serverless Container（如 AWS App Runner / Cloud Run / Fargate “无服务器运维”）**

- 适合：想保留原生 `listen` 方式，又希望“免运维/弹性”
- 代价：严格意义不是“函数式 serverless”，但改动最小；SQLite 仍不建议作为持久存储

**路线 C：保留 SQLite，但上持久卷（如 EFS）**

- 可行但不推荐：复杂度高、性能/并发/锁问题明显，且破坏 Serverless 简洁性

---

## 1. 需求澄清（必须先定方向）

在实施前，需要明确“Serverless”的目标平台与约束。推荐先让需求方确认以下问题：

- 目标平台：AWS Lambda / 阿里云函数计算 / 腾讯云 SCF / Cloudflare Workers（Go 不友好）/ GCP Cloud Functions
- 触发方式：HTTP API（API Gateway/网关）还是定时任务/队列事件
- 数据持久化：是否允许引入外部数据库（RDS/MySQL/Postgres）或云托管 KV
- 加密接口的返回格式：是否允许继续返回 `[]byte`（加密二进制）还是必须 base64/JSON

如果无法确认，先按“路线 A：AWS Lambda + API Gateway + 外部数据库”做参考实现，因为生态最成熟。

---

## 2. 代码结构改造方案（保持双运行模式）

### 2.1 目标结构

新增/调整文件建议如下（尽量少动现有文件）：

**新增**

- `cmd/lambda/main.go`：Lambda 入口（handler）
- `internal/bootstrap/bootstrap.go`：统一初始化（配置、DB、router）
- `internal/pkg/database/connect.go`：更清晰的 DB 初始化与复用（可拆分）
- `docs/serverless.md`：运行与部署文档

**修改**

- `cmd/main.go`：改为调用 bootstrap（不直接散落初始化逻辑）
- `internal/pkg/database/db.go`：支持“禁用 AutoMigrate / 支持多驱动 / 连接复用策略”
- `internal/config/config.go`：扩展配置（driver、migrate 开关等）

### 2.2 改造原则

- **路由只构建一次**：利用冷启动阶段构建 `gin.Engine`，热启动复用
- **DB 连接尽量复用**：将 `database.DB` 作为进程级单例（Lambda 容器复用时可以复用）
- **迁移不要默认开启**：避免 cold start 时做 migration 造成延迟或并发竞态

---

## 3. Task 分解（带可直接落地的步骤）

### Task 1: 增强配置以支持 Serverless 与多数据库

**Files:**
- Modify: [config.go](file:///workspace/internal/config/config.go#L1-L36)
- Modify: [config_demo.yaml](file:///workspace/configs/config_demo.yaml#L1-L10)
- Modify: [config.yaml](file:///workspace/config.yaml#L1-L5)

- [ ] **Step 1: 扩展 Config 结构体（DB driver、migrate 开关、serverless 标志）**

```go
// internal/config/config.go
type Config struct {
	Server struct {
		Port         int           `mapstructure:"port"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout"`
	}
	Database struct {
		Driver          string        `mapstructure:"driver"`           // 例如：sqlite / mysql / postgres
		DSN             string        `mapstructure:"dsn"`
		AutoMigrate     bool          `mapstructure:"auto_migrate"`     // 生产环境/Serverless 一般关闭
		MaxIdleConns    int           `mapstructure:"max_idle_conns"`
		MaxOpenConns    int           `mapstructure:"max_open_conns"`
		ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	}
	Serverless struct {
		Enabled bool `mapstructure:"enabled"` // 仅用于区分运行模式/默认策略
	}
}
```

- [ ] **Step 2: 设置合理默认值**

```go
// internal/config/config.go
viper.SetDefault("database.driver", "sqlite")
viper.SetDefault("database.auto_migrate", true)
viper.SetDefault("serverless.enabled", false)
```

- [ ] **Step 3: 更新配置示例**

```yaml
# configs/config_demo.yaml
server:
  port: 8080
  read_timeout: 15s
  write_timeout: 15s

database:
  driver: sqlite
  dsn: "file:data.db?cache=shared&_fk=1&_journal_mode=WAL"
  auto_migrate: true

serverless:
  enabled: false
```

- [ ] **Step 4: 运行 `go test ./...` 验证仅改配置不破坏编译**

Run: `go test ./...`
Expected: PASS

---

### Task 2: 重构 DB 初始化以适配 Serverless（迁移可控 + 连接复用）

**Files:**
- Modify: [db.go](file:///workspace/internal/pkg/database/db.go#L1-L34)
- Create: `internal/pkg/database/options.go`
- Test: `internal/pkg/database/db_test.go`

- [ ] **Step 1: 引入 Options，避免参数爆炸**

```go
// internal/pkg/database/options.go
package database

import "time"

type Options struct {
	Driver          string
	DSN             string
	AutoMigrate     bool
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}
```

- [ ] **Step 2: 将 Connect(dsn) 改造为 Connect(opts)**

```go
// internal/pkg/database/db.go
package database

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"stt/internal/model"
)

var DB *gorm.DB

func Connect(opts Options) error {
	if DB != nil {
		return nil // 关键：复用连接（Lambda 容器复用时收益明显）
	}

	if opts.Driver == "" {
		opts.Driver = "sqlite"
	}
	if opts.MaxOpenConns == 0 {
		opts.MaxOpenConns = 1 // SQLite 写入限制：单连接更安全
	}
	if opts.MaxIdleConns == 0 {
		opts.MaxIdleConns = 5
	}
	if opts.ConnMaxLifetime == 0 {
		opts.ConnMaxLifetime = time.Hour
	}

	var (
		db  *gorm.DB
		err error
	)

	switch opts.Driver {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(opts.DSN), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: false,
		})
	default:
		return fmt.Errorf("unsupported database driver: %s", opts.Driver)
	}

	if err != nil {
		return fmt.Errorf("failed to connect database: %v", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(opts.MaxOpenConns)
	sqlDB.SetMaxIdleConns(opts.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(opts.ConnMaxLifetime)

	if opts.AutoMigrate {
		if err := db.AutoMigrate(&model.StockMonitor{}); err != nil {
			return err
		}
	}

	DB = db
	return nil
}
```

- [ ] **Step 3: 单测（至少验证 AutoMigrate 开关不 panic）**

```go
// internal/pkg/database/db_test.go
package database

import "testing"

func TestConnect_AutoMigrateDisabled(t *testing.T) {
	DB = nil
	err := Connect(Options{
		Driver:      "sqlite",
		DSN:         "file::memory:?cache=shared",
		AutoMigrate: false,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if DB == nil {
		t.Fatalf("expected DB to be initialized")
	}
}
```

- [ ] **Step 4: 运行 `go test ./...`**

Run: `go test ./...`
Expected: PASS

---

### Task 3: 引入统一 bootstrap（本地与 Serverless 复用初始化）

**Files:**
- Create: `internal/bootstrap/bootstrap.go`
- Modify: [main.go](file:///workspace/cmd/main.go#L1-L30)
- Test: `internal/bootstrap/bootstrap_test.go`

- [ ] **Step 1: 编写 Bootstrap（返回 router）**

```go
// internal/bootstrap/bootstrap.go
package bootstrap

import (
	"stt/internal/config"
	"stt/internal/pkg/database"
	"stt/internal/route"
)

type App struct {
	Config *config.Config
	Router any // 这里用 any，避免 bootstrap 强依赖 gin 类型；调用方做类型断言
}

func Init(configPath string) (*App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	err = database.Connect(database.Options{
		Driver:          cfg.Database.Driver,
		DSN:             cfg.Database.DSN,
		AutoMigrate:     cfg.Database.AutoMigrate,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	if err != nil {
		return nil, err
	}

	r := route.SetupRouter()

	return &App{
		Config: cfg,
		Router: r,
	}, nil
}
```

- [ ] **Step 2: 本地入口改为调用 bootstrap**

```go
// cmd/main.go
package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"stt/internal/bootstrap"
)

func main() {
	app, err := bootstrap.Init("./config.yaml")
	if err != nil {
		log.Fatalf("bootstrap init failed: %v", err)
	}

	r, ok := app.Router.(*gin.Engine)
	if !ok {
		log.Fatalf("router type mismatch")
	}

	if err := r.Run(fmt.Sprintf(":%d", app.Config.Server.Port)); err != nil {
		log.Fatalf("server startup failed: %v", err)
	}
}
```

- [ ] **Step 3: 运行 `go test ./...`**

Run: `go test ./...`
Expected: PASS

---

### Task 4: 增加 AWS Lambda 入口（参考实现）

**适用前提：** 选择“路线 A：AWS Lambda + API Gateway”。

**Files:**
- Create: `cmd/lambda/main.go`
- Modify: `go.mod`（新增依赖）
- Create: `docs/serverless.md`

- [ ] **Step 1: 引入依赖**

Run:

```bash
go get github.com/aws/aws-lambda-go@latest
go get github.com/awslabs/aws-lambda-go-api-proxy/gin@latest
go mod tidy
```

Expected: `go.mod/go.sum` 更新

- [ ] **Step 2: 编写 Lambda 入口**

```go
// cmd/lambda/main.go
package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"stt/internal/bootstrap"
)

var ginLambda *ginadapter.GinLambda

func init() {
	// 关键：init 在 cold start 时执行一次，尽量在这里完成重资源初始化
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config.yaml"
	}

	app, err := bootstrap.Init(configPath)
	if err != nil {
		log.Fatalf("bootstrap init failed: %v", err)
	}

	r, ok := app.Router.(*gin.Engine)
	if !ok {
		log.Fatalf("router type mismatch")
	}

	ginLambda = ginadapter.New(r)
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// 关键：每次请求走 Proxy，复用 gin.Router 与中间件链路
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(handler)
}
```

- [ ] **Step 3: 补充 Serverless 文档**

`docs/serverless.md` 至少包含：
1) 构建方式（zip 或 Lambda container image）  
2) 环境变量（CONFIG_PATH、DB_DSN 等）  
3) API Gateway 配置（proxy integration）  

- [ ] **Step 4: 运行编译验证**

Run:

```bash
go test ./...
go build ./cmd/lambda
```

Expected: PASS

---

### Task 5: 数据持久化策略落地（Serverless 必选）

**结论先行：** 如果要在函数中写入 `stock_monitors`（以及后续可能新增的业务表），就必须把 SQLite 换成“外部 DB”或“托管存储”。

此 Task 分两条支线（二选一），建议优先 **MySQL/Postgres**：

#### 方案 5A：引入 MySQL/Postgres（推荐）

**Files:**
- Modify: `internal/pkg/database/db.go`（driver switch）
- Modify: `internal/config/config.go`（driver 默认策略）

- [ ] **Step 1: 选择并引入驱动**
  - MySQL：`gorm.io/driver/mysql`
  - Postgres：`gorm.io/driver/postgres`

- [ ] **Step 2: 扩展 Connect 的 switch**

```go
// internal/pkg/database/db.go（示意，需与现有代码合并）
switch opts.Driver {
case "sqlite":
	// ...
case "mysql":
	// 使用 mysql.Open(opts.DSN)
case "postgres":
	// 使用 postgres.Open(opts.DSN)
default:
	return fmt.Errorf("unsupported database driver: %s", opts.Driver)
}
```

#### 方案 5B：继续 SQLite，但改为内存/临时文件（仅用于只读或演示）

**说明：** 该方案会丢数据，仅适合 demo 或只读场景。

- [ ] **Step 1: 在 serverless 配置中使用内存 DSN**

```yaml
database:
  driver: sqlite
  dsn: "file::memory:?cache=shared"
  auto_migrate: true
```

---

## 4. 验证清单（Definition of Done）

- [ ] 本地模式 `go run cmd/main.go` 正常启动，股票相关 API（`/api/v1/std/*`）可用
- [ ] Serverless 模式（以 AWS Lambda 为例）可通过 `go build ./cmd/lambda` 编译
- [ ] DB 持久化策略明确：生产不依赖本地 `data.db`
- [ ] `go test ./...` 通过（至少有 bootstrap/database 的基础单测）
- [ ] 文档补充：`docs/serverless.md` 说明部署与配置

---

## 5. 自检（Spec coverage / Placeholder scan）

- 覆盖点：
  - 评估可行性：0、1 章节
  - 架构改造：2、3 章节
  - 运行方式：Task 4 + DoD
  - 依赖与限制：0.2、Task 5
- Placeholder 扫描：
  - 所有任务均给出明确文件路径与可落地代码/命令；仅“方案选择”以分支方式呈现，属于需求决策点，不是实现占位符
