# TechMemo Backend

基于 Gin + GORM 的后端 API 服务

## 项目结构

```
backend/
├── main.go              # 入口文件：初始化配置、数据库、AI Worker，启动服务
├── config/              # 配置管理
│   ├── config.go        # 配置结构和加载（Viper）
│   └── config.yaml      # 配置文件
├── bootstrap/           # 依赖注入：初始化并组装所有 DAO / Service
│   └── app.go
├── database/            # 数据库连接和初始化（GORM + pgvector）
│   └── database.go
├── model/               # 数据模型（GORM Gen 自动生成，勿手动修改）
│   ├── user.gen.go
│   ├── note.gen.go
│   ├── note_version.gen.go
│   ├── note_tag.gen.go
│   ├── category.gen.go
│   ├── tag.gen.go
│   ├── knowledge_point.gen.go
│   ├── knowledge_relation.gen.go
│   ├── embedding.gen.go
│   ├── ai_process_log.gen.go
│   └── embedding_custom.go  # 手写的向量相关扩展
├── query/               # GORM Gen 生成的类型安全查询层（勿手动修改）
├── dao/                 # 数据访问层：封装数据库操作
│   ├── user.go
│   ├── note.go
│   ├── category.go
│   ├── tag.go
│   ├── ai.go
│   ├── knowledge_point.go
│   └── params.go        # 通用查询参数
├── service/             # 业务逻辑层
│   ├── user.go
│   ├── note.go
│   ├── category.go
│   ├── tag.go
│   ├── ai.go
│   └── knowledge_point.go
├── handler/             # HTTP 处理层（控制器）
│   ├── user.go
│   ├── note.go
│   ├── category.go
│   ├── tag.go
│   ├── ai.go
│   ├── knowledge_point.go
│   └── dto/             # 请求/响应数据结构
│       ├── user.go
│       ├── note.go
│       ├── category.go
│       ├── tag.go
│       ├── ai.go
│       └── knowledge_point.go
├── ai/                  # AI 处理模块
│   ├── client/          # AI 客户端（OpenAI 兼容接口）
│   │   ├── client.go    # 接口定义 + OpenAI 实现（chat/embedding 分离配置）
│   │   └── types.go     # 知识点等数据结构
│   ├── queue/           # 异步任务队列（内存实现）
│   │   └── queue.go
│   └── worker/          # 队列消费 Worker
│       ├── worker.go
│       ├── handler.go   # 任务处理逻辑（知识抽取、向量化）
│       └── queue.go
├── router/              # 路由注册
│   └── router.go
├── middleware/          # 中间件
│   ├── cors.go          # CORS 跨域
│   ├── logger.go        # 日志
│   └── jwt.go           # JWT 认证
├── common/              # 公共工具
│   ├── response/        # 统一响应格式
│   └── errors/          # 自定义错误
├── utils/               # 工具函数
│   ├── jwt.go           # JWT 生成与解析
│   └── password.go      # 密码加密
├── cmd/gen/             # 代码生成工具
│   └── main.go          # 运行 GORM Gen
└── docs/                # Swagger 文档（自动生成）
```

## 分层架构

请求流转路径：

```
HTTP Request
    → middleware（JWT 认证、日志、CORS）
    → handler（参数绑定与校验）
    → service（业务逻辑）
    → dao（数据访问，封装 query 层）
    → query（GORM Gen 生成的类型安全查询）
    → PostgreSQL
```

AI 异步处理路径：

```
handler.POST /ai/note/:id
    → AIService.EnqueueTask（入队）
    → Worker（消费队列）
    → ai/client（调用 LLM：分类 → 知识抽取 → 向量化）
    → dao（写入 knowledge_point / embedding / ai_process_log）
```

## 快速开始

### 1. 安装依赖

```bash
cd backend
go mod tidy
```

### 2. 配置文件

编辑 `config/config.yaml`：

```yaml
server:
  port: "8080"
  mode: debug  # debug / release

database:
  host: localhost
  port: "5432"
  user: postgres
  password: your-password
  dbname: techmemo
  sslmode: disable

jwt:
  secret: your-secret-key
  expire_hour: 24

ai:
  chat:
    provider: deepseek   # openai / deepseek 等 OpenAI 兼容接口
    api_key: your-api-key
    model: deepseek-chat
    base_url: https://api.deepseek.com/v1
  embedding:
    provider: openai
    api_key: your-api-key
    model: text-embedding-3-small
    base_url: https://api.openai.com/v1
```

> chat 和 embedding 支持分别配置不同的 provider，可以混用。

### 3. 运行

```bash
go run main.go
```

服务启动在 `http://localhost:8080`

### 4. 查看 API 文档

```
http://localhost:8080/swagger-ui/index.html
```

## 主要依赖

- **Gin**: Web 框架
- **GORM + GORM Gen**: ORM + 类型安全查询代码生成
- **PostgreSQL + pgvector**: 数据库 + 向量检索扩展
- **go-openai**: OpenAI 兼容 AI 客户端
- **JWT**: 身份认证
- **Viper**: 配置管理
- **Swagger**: API 文档

## 开发指南

### 添加新的 API 接口

1. 在 `handler/dto/` 中定义请求/响应结构体
2. 在 `dao/` 中添加数据库操作方法
3. 在 `service/` 中实现业务逻辑
4. 在 `handler/` 中实现 HTTP 处理函数，添加 Swagger 注释
5. 在 `router/router.go` 中注册路由
6. 运行 `swag init` 更新文档

### 更新数据模型

修改数据库 schema 后，重新运行代码生成：

```bash
go run cmd/gen/main.go
```

这会重新生成 `model/*.gen.go` 和 `query/*.gen.go`，不要手动编辑这些文件。

### Swagger 注释示例

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
    // ...
}
```

## API 概览

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /api/v1/auth/register | 用户注册 | 否 |
| POST | /api/v1/auth/login | 用户登录 | 否 |
| GET  | /api/v1/auth/profile | 获取当前用户信息 | 是 |
| GET  | /api/v1/categories | 获取分类列表 | 是 |
| POST | /api/v1/categories | 创建分类 | 是 |
| PUT  | /api/v1/categories/:id | 更新分类 | 是 |
| DELETE | /api/v1/categories/:id | 删除分类 | 是 |
| GET  | /api/v1/tags | 获取标签列表 | 是 |
| POST | /api/v1/tags | 创建标签 | 是 |
| PUT  | /api/v1/tags/:id | 更新标签 | 是 |
| DELETE | /api/v1/tags/:id | 删除标签 | 是 |
| GET  | /api/v1/notes | 获取笔记列表 | 是 |
| POST | /api/v1/notes | 创建笔记 | 是 |
| GET  | /api/v1/notes/:id | 获取笔记详情 | 是 |
| PUT  | /api/v1/notes/:id | 更新笔记 | 是 |
| PUT  | /api/v1/notes/:id/tags | 更新笔记标签 | 是 |
| DELETE | /api/v1/notes/:id | 删除笔记 | 是 |
| GET  | /api/v1/notes/:id/versions | 获取版本历史 | 是 |
| POST | /api/v1/notes/:id/versions/:version_id/restore | 恢复历史版本 | 是 |
| GET  | /api/v1/knowledge-points | 获取知识点列表 | 是 |
| GET  | /api/v1/knowledge-points/:id | 获取知识点详情 | 是 |
| PUT  | /api/v1/knowledge-points/:id | 更新知识点 | 是 |
| DELETE | /api/v1/knowledge-points/:id | 删除知识点 | 是 |
| POST | /api/v1/ai/note/:id | 触发笔记 AI 处理 | 是 |
| GET  | /api/v1/ai/note/:id/status | 获取 AI 处理状态 | 是 |

### 待实现接口

- `POST /api/v1/mindmap/generate` — 生成思维导图
- `GET  /api/v1/mindmap/global` — 全局知识图谱
- `POST /api/v1/search/semantic` — 语义搜索
- `GET  /api/v1/search/keyword` — 关键词搜索
- `POST /api/v1/qa/ask` — 智能问答
- `GET  /api/v1/stats/overview` — 统计概览
- `GET  /api/v1/stats/categories` — 分类统计

## 健康检查

```bash
curl http://localhost:8080/health
```
