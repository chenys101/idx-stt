# STT 项目 Code Wiki

## 1. 项目概览

STT 是一个基于 Go 的轻量 HTTP 服务，提供两类核心能力：

- 用户管理：创建用户、按 ID 查询用户、列出用户
- 股票数据：从新浪行情接口拉取股票数据，结合本地「监控配置」(StockMonitor) 生成返回结果；其中一条接口会对结果进行混淆+加密

服务默认使用 SQLite 作为持久化存储，采用 Gin 提供 HTTP 路由与中间件机制。

## 2. 技术栈与依赖

**语言与构建**

- Go `1.24.0`（模块名 `stt`）：[go.mod](file:///workspace/go.mod#L1-L14)
- 运行：`go run cmd/main.go`（入口）：[main.go](file:///workspace/cmd/main.go#L1-L30)
- 镜像：多阶段构建 Dockerfile：[Dockerfile](file:///workspace/Dockerfile#L1-L26)

**关键三方库**

- Web：`github.com/gin-gonic/gin`、`github.com/gin-contrib/cors`：[go.mod](file:///workspace/go.mod#L6-L13)
- 配置：`github.com/spf13/viper`：[config.go](file:///workspace/internal/config/config.go#L1-L36)
- 数据库/ORM：`gorm.io/gorm` + `github.com/glebarez/sqlite`：[db.go](file:///workspace/internal/pkg/database/db.go#L1-L34)
- 编码转换：`golang.org/x/text`（用于 GB18030 转 UTF-8）：[sina.go](file:///workspace/internal/pkg/stock/sina.go#L11-L13)

## 3. 仓库结构

```text
/workspace
  cmd/                       # 启动入口
  internal/
    config/                  # 配置加载（Viper）
    route/                   # 路由注册（Gin）
    controller/              # HTTP Handler（User/Stock）
    middleware/              # Gin 中间件（内容协商）
    model/                   # GORM 模型（User/StockMonitor 等）
    pkg/                     # 可复用包（db/stock/encrypt/response 等）
  configs/                   # 配置示例
  config.yaml                # 默认运行配置
  data.db                    # SQLite 数据文件（运行时使用）
  Dockerfile                 # 镜像构建
  .github/workflows/         # CI（Docker build/push）
```

## 4. 整体架构

### 4.1 分层与调用方向

- **启动层**：`cmd/main.go`
- **基础设施层**：`internal/config`、`internal/pkg/database`、`internal/pkg/app`
- **Web 层**：`internal/route`、`internal/middleware`、`internal/controller`
- **领域/业务支撑**：`internal/pkg/stock`、`internal/pkg/encrypt`、`internal/pkg/strutil`
- **数据模型层**：`internal/model`

调用方向（由上至下）：

```text
cmd/main.go
  -> config.Load
  -> database.Connect
  -> route.SetupRouter
        -> middleware.ContentNegotiation
        -> controller.* (handlers)
              -> database.DB (GORM)
              -> stock.GetStockData (外部行情)
              -> encrypt.NewEncryptor + Encrypt
              -> app.Success / app.AbortWithError
```

### 4.2 启动与生命周期

入口函数见：[main.go](file:///workspace/cmd/main.go#L11-L29)

1. 读取配置文件 `./config.yaml`：`config.Load("./config.yaml")`  
2. 初始化数据库连接并执行 AutoMigrate：`database.Connect(cfg.Database.DSN)`  
3. 注册路由与中间件：`route.SetupRouter()`  
4. 监听端口启动服务：`r.Run(fmt.Sprintf(":%d", cfg.Server.Port))`  

## 5. 主要模块职责（按目录）

### 5.1 cmd/

- 启动入口：[main.go](file:///workspace/cmd/main.go#L1-L30)

### 5.2 internal/config

- 配置结构体：`Config`：[config.go](file:///workspace/internal/config/config.go#L8-L20)
- 加载逻辑：`Load(configPath string)` 使用 Viper 读取并 Unmarshal：[config.go](file:///workspace/internal/config/config.go#L22-L36)

### 5.3 internal/pkg/database

- 全局 DB：`var DB *gorm.DB`：[db.go](file:///workspace/internal/pkg/database/db.go#L10-L10)
- 初始化：`Connect(dsn string)`  
  - 打开 SQLite
  - 设置连接池（SQLite 仅 1 个 open conn）
  - AutoMigrate：`model.User`、`model.StockMonitor`  
  见：[db.go](file:///workspace/internal/pkg/database/db.go#L12-L34)

### 5.4 internal/route

- `SetupRouter()`：创建 Gin 引擎、配置 CORS、挂载内容协商中间件、注册 `/api/v1` 下的路由  
  见：[router.go](file:///workspace/internal/route/router.go#L10-L40)

### 5.5 internal/middleware

- `ContentNegotiation()`：从请求头推断输入/输出格式，写入 Gin Context 键：
  - `input_format`：从 `Content-Type` 判断 `json` / `text`
  - `output_format`：从 `Accept` 判断 `json` / `text`  
  见：[negotiation.go](file:///workspace/internal/middleware/negotiation.go#L8-L33)

### 5.6 internal/pkg/app

- `Success(c, code, data)`：按 `output_format` 返回 JSON 或纯文本：[response.go](file:///workspace/internal/pkg/app/response.go#L13-L27)
- `AbortWithError(c, code, msg)`：统一错误返回：[response.go](file:///workspace/internal/pkg/app/response.go#L29-L43)

### 5.7 internal/model

- 用户表：`model.User`：[user.go](file:///workspace/internal/model/user.go#L9-L14)
- 股票监控表：`model.StockMonitor`：[stock.go](file:///workspace/internal/model/stock.go#L15-L20)
- 股票返回 DTO：
  - `model.StockBase`（明文返回）：[stock.go](file:///workspace/internal/model/stock.go#L5-L9)
  - `model.StockBaseEn`（混淆/拼接后的结构）：[stock.go](file:///workspace/internal/model/stock.go#L10-L14)

### 5.8 internal/controller

**UserController**

- `CreateUser`：支持 JSON 或 text/plain 的输入解析；写入 SQLite：[user.go](file:///workspace/internal/controller/user.go#L14-L51)
- `GetUser`：按 `:id` 查询：[user.go](file:///workspace/internal/controller/user.go#L52-L72)
- `ListUsers`：列表查询：[user.go](file:///workspace/internal/controller/user.go#L74-L81)

**StockController**

- 数据获取入口：
  - `GetStockData`：返回加密结果（见「6.2」）：[stock.go](file:///workspace/internal/controller/stock.go#L19-L79)
  - `GetStockDataNotEncrypt`：返回结构化明文结果：[stock.go](file:///workspace/internal/controller/stock.go#L81-L120)
- 监控配置 CRUD：
  - `CreateStockMonitor`：[stock.go](file:///workspace/internal/controller/stock.go#L122-L140)
  - `UpdateStockMonitor`：[stock.go](file:///workspace/internal/controller/stock.go#L142-L171)
  - `GetAllStockMonitors`：[stock.go](file:///workspace/internal/controller/stock.go#L173-L182)

### 5.9 internal/pkg/stock（外部行情适配）

- `GetStockData(StockCodes ...string)`：拼接新浪接口 URL → HTTP 请求 → GB18030 解码 → 分行解析 → 计算涨跌幅/涨跌额  
  见：[sina.go](file:///workspace/internal/pkg/stock/sina.go#L77-L122)
- `ParseFullSingleStockData(data string)`：解析单行返回为 `StockInfo`：[sina.go](file:///workspace/internal/pkg/stock/sina.go#L124-L208)
- `addStockFollowData(stockData *StockInfo)`：补充 `ChangePercent`、`ChangePrice`、`HighRate`、`LowRate`：[sina.go](file:///workspace/internal/pkg/stock/sina.go#L210-L254)

### 5.10 internal/pkg/encrypt（加密抽象）

- 抽象接口：`Encryptor`：[encryptor.go](file:///workspace/internal/pkg/encrypt/encryptor.go#L9-L13)
- 工厂：`NewEncryptor(algoMethod string, key any)`：[encryptor.go](file:///workspace/internal/pkg/encrypt/encryptor.go#L27-L44)
- AES 实现入口：`algo.NewAesEncryptor(key []byte)`：[algo.go](file:///workspace/internal/pkg/encrypt/algo/algo.go#L7-L10)

### 5.11 internal/pkg/strutil（混淆与数字编码）

- `AddNoise(input string)`：随机插入特殊字符，混淆字符串：[noise_util.go](file:///workspace/internal/pkg/strutil/noise_util.go#L63-L89)
- `FloatNumToChinese(num float64)`：把数值转换为带“扰动字符”的表示：[noise_util.go](file:///workspace/internal/pkg/strutil/noise_util.go#L12-L44)

## 6. 核心流程详解

### 6.1 HTTP 请求链路（通用）

以任意 API 为例（例如 `GET /api/v1/users`）：

1. Gin Router 匹配路由：[router.go](file:///workspace/internal/route/router.go#L19-L36)
2. 进入中间件 `ContentNegotiation()`，写入 `input_format/output_format`：[negotiation.go](file:///workspace/internal/middleware/negotiation.go#L8-L33)
3. 进入 Controller Handler，执行业务逻辑（DB/外部请求等）
4. 通过 `app.Success` 或 `app.AbortWithError` 做统一输出格式化：[response.go](file:///workspace/internal/pkg/app/response.go#L13-L43)

### 6.2 股票数据（加密接口）流程

处理函数：`StockController.GetStockData`：[stock.go](file:///workspace/internal/controller/stock.go#L19-L79)

1. 解析参数：优先取 query `list`，否则取 `code` 并从 `StockMonitor` 表中加载 `MonitorValue`
2. 调用 `stock.GetStockData(list)` 拉取行情并计算涨跌相关字段：[sina.go](file:///workspace/internal/pkg/stock/sina.go#L77-L122)
3. 将返回的 `StockInfo` 转换为 `model.StockBaseEn`，并对股票名、涨跌信息做混淆拼接：[stock.go](file:///workspace/internal/controller/stock.go#L46-L59)
4. JSON 序列化后执行 AES 加密：`encrypt.NewEncryptor(encrypt.AES256, key).Encrypt(...)`：[stock.go](file:///workspace/internal/controller/stock.go#L60-L76)
5. 通过 `app.Success` 返回加密后的 `[]byte` 数据：[stock.go](file:///workspace/internal/controller/stock.go#L78-L79)

补充说明：

- 代码中 AES 常量命名为 `aes-256-gcm`（[encryptor.go](file:///workspace/internal/pkg/encrypt/encryptor.go#L16-L19)），但当前 `aesEncryptor.Encrypt` 实现使用 ECB + PKCS7（[aes.go](file:///workspace/internal/pkg/encrypt/algo/aes.go#L15-L25)）。如果未来需要解密对称闭环或对外对齐算法名称，建议优先核对该实现与常量/注释一致性。

## 7. API 清单（v1）

路由定义统一在：[router.go](file:///workspace/internal/route/router.go#L19-L36)

### 7.1 Users

- `POST /api/v1/users` → `UserController.CreateUser`
  - JSON：`{"name":"xxx","email":"xxx@xx.com"}`
  - 或 text/plain：按 `Name:`/`Email:` 行解析：[user.go](file:///workspace/internal/controller/user.go#L20-L38)
- `GET /api/v1/users/:id` → `UserController.GetUser`
- `GET /api/v1/users` → `UserController.ListUsers`

### 7.2 Stock（std）

- `GET /api/v1/std` → `StockController.GetStockData`（加密）
  - query：`list=sh000001,sz000001` 或 `code=sg`（从监控表取 list）
- `GET /api/v1/std/nc` → `StockController.GetStockDataNotEncrypt`（明文）
- `POST /api/v1/std/createStdMonitor` → `StockController.CreateStockMonitor`
  - JSON：`{"code":"sg","monitorValue":"shxxx,szxxx"}`
- `POST /api/v1/std/updateStdMonitor` → `StockController.UpdateStockMonitor`
- `GET /api/v1/std/getAllStdMonitors` → `StockController.GetAllStockMonitors`

## 8. 配置与数据

### 8.1 配置文件

默认配置文件在仓库根目录：

- 当前运行配置：[config.yaml](file:///workspace/config.yaml#L1-L5)
- 示例配置（含更多字段）：[config_demo.yaml](file:///workspace/configs/config_demo.yaml#L1-L10)

`Config` 结构体字段映射见：[config.go](file:///workspace/internal/config/config.go#L8-L20)

### 8.2 SQLite 数据文件

默认使用 `data.db`（仓库已包含）。初始化时会自动创建/迁移表：

- `users`（由 `model.User` 映射）
- `stock_monitors`（由 `model.StockMonitor` 映射）

迁移逻辑见：[db.go](file:///workspace/internal/pkg/database/db.go#L25-L31)

## 9. 运行与构建

### 9.1 本地运行

README 中给出的最小步骤：[README.md](file:///workspace/README.md#L1-L8)

```bash
go mod tidy
cp configs/config_demo.yaml config.yaml
go run cmd/main.go
```

服务默认端口由 `config.yaml` 决定（当前为 `9002`）：[config.yaml](file:///workspace/config.yaml#L1-L5)

### 9.2 Docker 运行

- 构建与启动逻辑：[Dockerfile](file:///workspace/Dockerfile#L1-L26)
- 镜像默认暴露端口 `9002`（与仓库根配置一致）：[Dockerfile](file:///workspace/Dockerfile#L25-L26)

示例：

```bash
docker build -t idx-stt:local .
docker run --rm -p 9002:9002 idx-stt:local
```

### 9.3 CI（GitHub Actions）

- workflow：Build Docker Image：[docker-build.yml](file:///workspace/.github/workflows/docker-build.yml#L1-L40)
- 行为：只 build 不 push（`push: false`），并启用 GHA cache

## 10. 进一步阅读索引（入口导航）

- 启动入口：[main.go](file:///workspace/cmd/main.go#L1-L30)
- 路由定义：[router.go](file:///workspace/internal/route/router.go#L1-L40)
- User API：[user.go](file:///workspace/internal/controller/user.go#L1-L83)
- Stock API：[stock.go](file:///workspace/internal/controller/stock.go#L1-L182)
- 股票数据解析（新浪）：[sina.go](file:///workspace/internal/pkg/stock/sina.go#L1-L254)
- 配置加载：[config.go](file:///workspace/internal/config/config.go#L1-L36)
- DB 初始化：[db.go](file:///workspace/internal/pkg/database/db.go#L1-L34)

