# mail-service 开发指南

## 项目概述

基于 Go + Gin 的邮件发送服务，支持多邮箱轮询发送、Redis 缓存、MySQL 持久化，内置管理后台（SPA）。

## 技术栈

| 组件 | 技术 | 版本 |
|------|------|------|
| Web 框架 | Gin | v1 |
| 数据库 | MySQL | go-sql-driver/mysql |
| 缓存 | Redis | go-redis/v8 |
| 认证 | JWT | golang-jwt/jwt/v5 |
| 邮件 | gomail | gomail.v2 |
| 图表 | ECharts | 5 |
| CI/CD | Gitea Actions | — |

## 文件结构

| 文件 | 职责 | 修改影响范围 |
|------|------|-------------|
| `main.go` | 入口、服务初始化、邮件发送 Handler | 全局 |
| `admin.go` | 后台路由、JWT 认证、API 接口 | 后台 API、路由 |
| `models.go` | 数据模型、MySQL 表操作、HTML 感知脱敏 | 数据库、日志展示 |
| `stats.go` | Redis 数据统计（动态天数、域名分布） | 仪表盘数据 |
| `ban.go` | IP 登录封禁（指数退避） | 登录安全 |
| `templates.go` | 所有前端 HTML/CSS/JS（内联模板） | 后台 UI、页面交互 |
| `Dockerfile` | 多阶段构建镜像 | 构建产物 |
| `docker-compse.yml` | Docker Compose 部署配置 | 部署行为 |
| `.gitea/workflows/docker.yml` | CI/CD 流水线配置 | 自动构建、部署 |

## 核心设计

### 邮件发送流程
1. 从 Redis Lua 脚本轮询获取 `mailid`（1 到 MySQL 中邮箱总数循环）
2. 先查 Redis 缓存邮件配置，未命中再查 MySQL
3. 使用 gomail 发送邮件（SSL 465 端口），15s 超时
4. 异步双写：goroutine 写 Redis → 嵌套 goroutine 写 MySQL
5. 邮件内容脱敏：`sanitizeSensitive()` HTML 感知模式，只脱敏标签外文本

### 后台架构
- 单页应用（SPA）：`/admin/dashboard`、`/admin/mails`、`/admin/logs`、`/admin/templates`、`/admin/audit` 返回同一 HTML
- 视图通过 CSS `display` 切换，ECharts 实例只初始化一次
- JWT Cookie 认证，24h 过期
- IP 登录失败 3 次后指数退避封禁（5min * 2^(n-3)，上限 24h）
- SSE 实时推送：仪表盘自动刷新统计，日志页自动 prepend 新记录

### 数据流
```
发送请求 → Redis 轮询 mailid → 获取配置 → 发送邮件
                                    ↓
                              异步保存日志
                                    ↓
                         Redis (7天TTL) → MySQL (持久化)
```

## CI/CD 流水线

基于 Gitea Actions（`.gitea/workflows/docker.yml`）：

- **触发条件**：推送 `main` 分支或 `v*` 标签
- **构建**：Docker Buildx 多阶段构建，推送到腾讯云 CCR（`ccr.ccs.tencentyun.com/iskycc/mail-service`）
- **标签**：每次推送生成 `latest`、`sha`、`ref_name` 三个标签
- **部署**：SSH 到远程服务器执行 `docker compose pull && up -d`
- **认证**：
  - 登录腾讯云 CCR 用于推送（`DOCKERHUB_USERNAME` / `DOCKERHUB_TOKEN`）
  - 登录 Docker Hub 用于拉取基础镜像（`DOCKERHUB_PULL_USERNAME` / `DOCKERHUB_PULL_PASSWORD`）

## 环境变量

```bash
# Redis
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PWD=

# MySQL
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PWD=
MYSQL_DBNAME=mail

# Admin
ADMIN_USER=admin
ADMIN_PASSWORD=changeme
JWT_SECRET=your-secret-key

# 反向代理
TRUSTED_PROXIES=your-proxy-ip  # 配置信任的代理 IP，多个用逗号分隔
```

## 开发规范

### 代码规范
- **每次提交前必须更新版本号**：修改 `main.go` 顶部的 `version` 常量（如 `"1.1.0"`），这是唯一确认部署是否生效的方式
- MySQL 建表使用 `ensureTable()` 先检查再创建，不删除已有表
- 新增后台 API 需在 `admin.go` 中注册路由，并在 `templates.go` 前端调用
- 修改数据模型需同步检查 `ensureTable()` 中的建表语句

### 前端规范
- **不要**在 `templates.go` 的 JS 中使用反引号模板字符串（Go raw string 会冲突）
- **不要**修改 `sharedDarkCSS` 中的变量名，所有页面共享
- 新增后台页面需加入 SPA 视图体系（`showView()` 函数）
- ECharts 主题切换需 `dispose()` 后重新 `init()`
- 后台页面需确保 `showView()` 中加载对应数据（首次进入 + 缓存路径都要调用）
- ECharts 图表 resize 事件已绑定 `window`，窗口变化自动适配

### 安全规范
- 发信日志中的敏感信息通过 `sanitizeSensitive()` 脱敏，**不会**破坏 HTML 标签
- 前端拼接 HTML 时**必须**使用 `escapeHtml()` 函数转义用户输入，防止 XSS
- JWT Secret 生产环境务必设置为强随机字符串
- 后台登录密码不支持修改，需通过环境变量配置

### 部署规范
- 部署服务器需提前 `docker login ccr.ccs.tencentyun.com`
- `docker-compse.yml` 文件名缺少字母 `o`，是历史遗留，修改需谨慎评估影响

## 测试

```bash
# 运行全部测试
go test ./...

# 运行并输出详细日志
go test -v ./...
```

## 常用操作

```bash
# 编译
go build -o mail-service .

# 运行（需先设置环境变量）
export REDIS_HOST=127.0.0.1 REDIS_PORT=6379 ...
./mail-service

# 测试
go test ./...

# Docker 构建
docker build -t mail-service .

# Docker Compose
docker-compose -f docker-compse.yml up -d
```

## 端口

- 服务端口: `:22125`
- Admin 路径: `/admin/*`
- 发信接口: `GET/POST /`
