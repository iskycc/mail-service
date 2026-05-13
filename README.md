# mail-service

基于 Go + Gin 的邮件发送服务，支持多邮箱轮询、Redis 缓存、MySQL 持久化，内置可视化后台管理。

## 功能特性

- **多邮箱轮询发送**：支持配置多个 SMTP 邮箱，Redis Lua 脚本动态轮询切换
- **统一发信接口**：同时支持 GET、POST（form/json）请求，可选模板 ID 发送
- **Redis 缓存**：邮箱配置缓存 7 天，减少 MySQL 查询
- **MySQL 持久化**：自动建表，日志异步双写（Redis → MySQL）
- **日志脱敏**：自动替换邮件内容中的敏感信息（验证码）为 `****`，HTML 感知不破坏标签结构
- **请求追踪**：每次请求生成唯一 `request_id`，全链路日志标记便于排查
- **管理后台**：单页应用，支持仪表盘统计、邮箱池 CRUD、发信日志搜索、邮件模板管理、操作审计
- **暗黑模式**：全局主题切换，localStorage 持久化
- **SSE 实时推送**：仪表盘/日志页自动刷新新数据
- **IP 封禁**：登录失败 3 次后指数退避封禁（5min ~ 24h）
- **移动端适配**：登录页居中、顶栏导航可横向滚动、表格可左右滑动、弹窗自适应
- **CI/CD 自动部署**：Gitea Actions 推送 main 分支自动构建镜像并部署

## 技术栈

- Go 1.24 + Gin
- MySQL 8.0
- Redis 7
- JWT Cookie 认证
- ECharts 5 数据可视化

## 快速开始

### 环境要求

- Go 1.24+
- MySQL 5.7+
- Redis 6.0+

### 环境变量

| 变量 | 必填 | 默认值 | 说明 |
|------|:----:|--------|------|
| `REDIS_HOST` | 是 | — | Redis 服务器地址 |
| `REDIS_PORT` | 是 | — | Redis 端口 |
| `REDIS_PWD` | 否 | — | Redis 密码 |
| `MYSQL_HOST` | 是 | — | MySQL 服务器地址 |
| `MYSQL_PORT` | 是 | — | MySQL 端口 |
| `MYSQL_USER` | 是 | — | MySQL 用户名 |
| `MYSQL_PWD` | 是 | — | MySQL 密码 |
| `MYSQL_DBNAME` | 是 | — | MySQL 数据库名 |
| `ADMIN_USER` | 否 | `admin` | 后台登录账号 |
| `ADMIN_PASSWORD` | 否 | `changeme` | 后台登录密码 |
| `JWT_SECRET` | 否 | — | JWT 签名密钥（生产环境务必设置） |
| `TRUSTED_PROXIES` | 否 | — | 信任的反向代理 IP，多个用逗号分隔 |

示例：

```bash
export REDIS_HOST=127.0.0.1 REDIS_PORT=6379 REDIS_PWD=
export MYSQL_HOST=127.0.0.1 MYSQL_PORT=3306 MYSQL_USER=root MYSQL_PWD=your-password MYSQL_DBNAME=mail
export ADMIN_USER=admin ADMIN_PASSWORD=changeme JWT_SECRET=your-secret-key
```

### 编译运行

```bash
go build -o mail-service .
./mail-service
```

服务将监听 `:22125`。

### Docker 部署

```bash
docker build -t mail-service .
docker run -d -p 22125:22125 \
  -e REDIS_HOST=redis \
  -e MYSQL_HOST=mysql \
  -e ADMIN_USER=admin \
  -e ADMIN_PASSWORD=changeme \
  mail-service
```

或使用 Docker Compose:

```bash
docker-compose -f docker-compse.yml up -d
```

> 注意：`docker-compse.yml` 文件名缺少字母 `o`，是历史遗留，引用时请使用实际文件名。

### CI/CD 自动部署

项目使用 Gitea Actions 实现自动构建与部署。推送 `main` 分支或 `v*` 标签时触发：

1. 构建 Docker 镜像并推送到腾讯云 CCR（`ccr.ccs.tencentyun.com`）
2. SSH 到部署服务器执行 `docker compose up -d`

#### 所需 Secrets

在 Gitea 仓库 **Settings → Secrets** 中配置：

| Secret | 说明 |
|--------|------|
| `DOCKERHUB_USERNAME` | 腾讯云 CCR 登录用户名 |
| `DOCKERHUB_TOKEN` | 腾讯云 CCR 登录密码 |
| `DOCKERHUB_PULL_USERNAME` | Docker Hub 用户名（用于拉取基础镜像） |
| `DOCKERHUB_PULL_PASSWORD` | Docker Hub 密码或 Access Token |
| `REMOTE_HOST` | 部署服务器 IP |
| `REMOTE_USER` | 部署服务器 SSH 用户名 |
| `REMOTE_PORT` | 部署服务器 SSH 端口（可选，默认 22） |
| `SSH_PASSWORD` | 部署服务器 SSH 密码 |
| `COMPOSE_DIR` | 部署服务器上 docker-compse.yml 所在目录 |

> 部署服务器需提前执行 `docker login ccr.ccs.tencentyun.com`，确保能拉取私有镜像。

## API 接口

### 发送邮件

```bash
# GET
curl "http://localhost:22125/?user=recipient@example.com&subject=标题&body=<h1>内容</h1>"

# POST form
curl -X POST -d "user=recipient@example.com&subject=标题&body=内容" http://localhost:22125/

# POST JSON
curl -X POST -H "Content-Type: application/json" \
  -d '{"user":"recipient@example.com","subject":"标题","body":"内容"}' \
  http://localhost:22125/

# 使用模板发送
curl "http://localhost:22125/?user=recipient@example.com&template_id=1&vars=%7B%22code%22%3A%22123456%22%7D"
```

参数：

| 参数 | 必填 | 说明 |
|------|:----:|------|
| `user` | 是 | 收件人邮箱 |
| `subject` | 条件 | 邮件主题；使用 `template_id` 时可为空，自动使用模板主题 |
| `body` | 条件 | HTML 内容；使用 `template_id` 时可为空，自动使用模板正文 |
| `altbody` | 否 | 纯文本内容（默认取 `body`） |
| `tname` | 否 | 发件人名称（默认 `Libv团队`） |
| `template_id` | 否 | 邮件模板 ID |
| `vars` | 条件 | JSON 格式模板变量，如 `{"code":"123456"}`；使用含变量的模板时必填 |

返回：

```json
{"success": true, "info": "Message has been sent", "request_id": "20260511120000-0001"}
```

### 健康检查

```bash
curl http://localhost:22125/health
```

仅允许内网访问，返回 MySQL/Redis 连接状态。

### 后台管理

访问 `http://localhost:22125/admin/login`，使用 `ADMIN_USER` / `ADMIN_PASSWORD` 登录。

后台功能：

| 页面 | 路径 | 功能 |
|------|------|------|
| 仪表盘 | `/admin/dashboard` | 发信统计（支持 1/7/30 天）、趋势图、域名占比饼图、单邮箱统计 |
| 邮箱池 | `/admin/mails` | 邮箱配置 CRUD，支持测试发送验证 SMTP |
| 发信日志 | `/admin/logs` | 关键词搜索、状态筛选、分页查看（可切换每页条数）、CSV 导出 |
| 模板管理 | `/admin/templates` | 邮件模板 CRUD，支持 HTML 预览，支持 `{{变量名}}` 占位符 |
| 操作审计 | `/admin/audit` | 管理员操作记录查看 |

### 后台 API（需 JWT 认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/api/stats?days=7` | 仪表盘统计数据 |
| GET | `/admin/api/mail-stats?days=7` | 单邮箱发信统计 |
| GET | `/admin/api/mails` | 邮箱列表 |
| POST | `/admin/api/mails` | 新增邮箱 |
| PUT | `/admin/api/mails/:id` | 修改邮箱 |
| DELETE | `/admin/api/mails/:id` | 删除邮箱 |
| POST | `/admin/api/mails/:id/test` | 测试发送 |
| GET | `/admin/api/logs?limit=20&offset=0` | 日志列表 |
| GET | `/admin/api/logs-count` | 日志总数 |
| GET | `/admin/api/logs/:id` | 日志详情 |
| GET | `/admin/api/export-logs` | 导出 CSV |
| GET | `/admin/api/templates` | 模板列表 |
| POST | `/admin/api/templates` | 新增模板 |
| PUT | `/admin/api/templates/:id` | 修改模板 |
| DELETE | `/admin/api/templates/:id` | 删除模板 |
| GET | `/admin/api/audit-logs` | 审计日志列表 |
| GET | `/admin/api/version` | 系统版本 |
| GET | `/admin/api/events` | SSE 实时推送 |

**版本管理**：每次提交前务必更新 `main.go` 中的 `version` 常量，以便通过后台页面底部或 `/admin/api/version` 确认部署是否生效。

## 项目结构

```
mail-service/
├── main.go              # 入口、服务初始化、邮件发送 Handler
├── admin.go             # 后台路由、JWT 认证、API 接口
├── models.go            # 数据模型、MySQL 操作、HTML 感知脱敏
├── stats.go             # Redis 数据统计（动态天数、域名分布）
├── ban.go               # IP 封禁逻辑（指数退避）
├── ban_test.go          # IP 封禁单元测试
├── models_test.go       # 模型单元测试
├── templates.go         # 前端模板（SPA，含所有 HTML/CSS/JS）
├── templates/           # 静态模板文件（如 verification-template.html）
├── Dockerfile           # 多阶段镜像构建
├── docker-compse.yml    # Docker Compose 部署
├── .gitea/workflows/    # Gitea Actions CI/CD 配置
│   └── docker.yml
├── go.mod               # Go 依赖
└── go.sum               # Go 依赖校验
```

## 测试

```bash
# 运行全部测试
go test ./...

# 运行并输出详细日志
go test -v ./...
```

## 开源协议

MIT
