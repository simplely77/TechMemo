# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在此代码仓库中工作时提供指导。

## 项目概述

TechMemo 是一个基于大语言模型的个人技术知识管理系统。它能自动分析技术笔记、提取知识点、生成思维导图，并基于用户积累的知识提供智能问答。

**技术栈：**
- 后端：Go 1.24+ 配合 Gin 框架
- 数据库：PostgreSQL 配合 pgvector 扩展用于向量存储
- 队列：Redis（开发环境可回退到内存队列）
- AI：DeepSeek API 用于对话，本地 Python embedding 服务（all-MiniLM-L6-v2）
- ORM：GORM 配合 GORM Gen 用于类型安全的查询生成

## 架构设计

### 分层结构

```
HTTP Request
  → Middleware (JWT, CORS, Logger)
  → Handler (参数绑定/校验)
  → Service (业务逻辑)
  → DAO (数据访问，封装 query 层)
  → Query (GORM Gen 生成，类型安全)
  → PostgreSQL
```

### AI 处理流程

```
POST /ai/note/:id
  → AIService.EnqueueTask (入队到 Redis/Memory)
  → Worker (消费队列)
  → AI Client (LLM: 分类 → 提取知识 → 向量化)
  → DAO (保存到 knowledge_point / embedding / ai_process_log)
```

AI worker 在 goroutine 中异步运行，从队列中处理任务。任务包括笔记分类、知识点提取和 embedding 生成。

## 常用命令

### 开发环境

```bash
# 在仓库根目录：仅依赖 或 全栈（需 .env 中 APP_AI_CHAT_API_KEY；改后端用 docker compose up -d --build api）
docker compose up -d

# 运行后端服务
cd backend
go run main.go

# 服务运行在 http://localhost:8080
# Swagger 文档在 http://localhost:8080/swagger-ui/index.html
```

### 代码生成

```bash
# 重新生成 GORM 模型和查询（数据库 schema 变更后）
cd backend
go run cmd/gen/main.go

# 重新生成 Swagger 文档（添加/修改 API handler 后）
cd backend
swag init
```

### 测试

```bash
# 运行测试
cd backend
go test ./...

# 健康检查
curl http://localhost:8080/health
```

## 关键约定

### GORM Gen 自动生成的代码

**绝对不要手动编辑这些目录：**
- `backend/model/*.gen.go` - 自动生成的模型文件
- `backend/query/` - 自动生成的类型安全查询层

这些文件由 `go run cmd/gen/main.go` 重新生成。只有 `backend/model/embedding_custom.go` 是手动维护的，用于自定义向量操作。

### 添加新的 API 端点

遵循以下工作流程：
1. 在 `backend/handler/dto/` 中定义请求/响应 DTO
2. 在 `backend/dao/` 中添加数据库操作
3. 在 `backend/service/` 中实现业务逻辑
4. 在 `backend/handler/` 中创建带 Swagger 注解的 handler
5. 在 `backend/router/router.go` 中注册路由
6. 运行 `swag init` 更新文档

### Swagger 注解

所有 handler 函数必须包含 Swagger 注解：

```go
// @Summary 创建笔记
// @Tags 笔记
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateNoteRequest true "笔记内容"
// @Success 200 {object} response.Response{data=dto.NoteResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/notes [post]
func HandlerCreateNote(svc *service.NoteService) gin.HandlerFunc {
    // 实现代码
}
```

### 错误处理

使用统一的错误处理系统：
- 自定义错误定义在 `backend/common/errors/errors.go`
- 统一响应格式在 `backend/common/response/response.go`
- 所有 API 响应遵循格式：`{code, message, data}`

### 身份认证

大多数端点需要 JWT 认证：
- 登录/注册端点返回 JWT token
- 受保护的端点需要 `Authorization: Bearer <token>` 请求头
- JWT 中间件验证 token 并提取 user_id
- JWT 配置在 `backend/config/config.yml`

## 配置管理

配置由 Viper 管理，从 `backend/config/config.yml` 加载：

```yaml
server:
  port: "8080"
  mode: debug  # debug / release

database:
  host: localhost
  port: "5432"
  # ... PostgreSQL 连接详情

redis:
  enabled: true  # 设为 false 使用内存队列
  host: localhost
  port: "6379"

ai:
  chat:
    provider: deepseek
    api_key: your-key
    model: deepseek-chat
    base_url: https://api.deepseek.com
  embedding:
    provider: local
    model: all-MiniLM-L6-v2
    base_url: http://localhost:5000
```

**重要：** chat 和 embedding 可以使用不同的 provider。embedding 服务作为独立的 Python 服务运行在 Docker 中。

## 数据库表结构

核心数据表：
- `user` - 用户账户
- `category` - 粗粒度笔记分类
- `tag` - 细粒度语义标签（与笔记多对多）
- `note` - 用户原始笔记（Markdown 格式）
- `note_version` - 笔记版本历史
- `knowledge_point` - AI 提取的知识单元
- `knowledge_relation` - 知识点之间的关联关系
- `embedding` - 笔记和知识点的向量表示（pgvector）
- `ai_process_log` - 所有 AI 操作的审计日志

`embedding` 表使用 pgvector 的 `vector(384)` 类型进行语义搜索。笔记（粗粒度）和知识点（细粒度）都会生成 embedding。

## 依赖注入

`bootstrap` 包通过适当的依赖注入初始化所有服务：

```go
// bootstrap/app.go
type App struct {
    UserService           *service.UserService
    CategoryService       *service.CategoryService
    TagService            *service.TagService
    NoteService           *service.NoteService
    AIService             *service.AIService
    KnowledgePointService *service.KnowledgePointService
}
```

服务在 `bootstrap.InitApp()` 中初始化，并通过路由设置传递给 handler。

## AI Worker 系统

AI worker 系统异步处理任务：

1. **队列接口**：对 Redis 和内存实现的抽象
2. **Worker**：从队列消费任务，支持可配置的并发数
3. **Handler**：处理每个任务（提取知识、生成 embedding）
4. **重试逻辑**：失败的任务会以指数退避方式重试

Worker 在 `main.go` 中作为 goroutine 启动，运行直到 context 取消。

## 常见陷阱

1. **不要编辑生成的代码**：`model/*.gen.go` 和 `query/` 中的文件会被重新生成
2. **运行 swag init**：修改 handler 的 Swagger 注解后
3. **使用类型安全查询**：利用 GORM Gen 的查询构建器而不是原始 SQL
4. **检查队列配置**：如果配置中 `redis.enabled: true`，Redis 必须运行
5. **Embedding 服务**：AI 处理需要 embedding 服务运行（在仓库根目录 `docker compose up` 会拉起）
6. **向量维度**：Embedding 服务使用 384 维向量（all-MiniLM-L6-v2），而不是 1536

## 项目状态

已实现功能：
- 用户认证（注册、登录、个人信息）
- 笔记 CRUD 及版本历史
- 分类和标签管理
- AI 处理（分类、知识提取、embedding）
- 知识点管理
- 异步 AI worker 系统

计划功能（见 backend/README.md）：
- 思维导图生成
- 语义搜索
- 智能问答（RAG）
- 统计仪表板
