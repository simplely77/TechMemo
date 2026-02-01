# TechMemo Backend

基于 Gin + GORM 的后端 API 服务

## 项目结构

```
backend/
├── main.go              # 入口文件
├── config/              # 配置管理
│   ├── config.go        # 配置结构和加载
│   └── config.yaml      # 配置文件
├── database/            # 数据库
│   └── database.go      # 数据库连接和初始化
├── models/              # 数据模型
│   ├── user.go
│   ├── category.go
│   ├── tag.go
│   ├── note.go
│   ├── knowledge.go
│   └── ai.go
├── router/              # 路由
│   └── router.go
├── middleware/          # 中间件
│   ├── cors.go          # CORS 跨域
│   ├── logger.go        # 日志
│   └── jwt.go           # JWT 认证
├── controllers/         # 控制器（TODO）
├── services/            # 业务逻辑层（TODO）
├── utils/               # 工具函数
│   ├── jwt.go           # JWT 工具
│   ├── response.go      # 响应格式化
│   └── password.go      # 密码加密
└── docs/                # Swagger 文档（自动生成）
```

## 快速开始

### 1. 安装依赖

```bash
cd backend
go mod tidy
```

### 2. 配置数据库

编辑 `config/config.yaml`：

```yaml
database:
  host: localhost
  port: "5432"
  user: postgres
  password: your-password
  dbname: techmemo
  sslmode: disable
```

### 3. 运行项目

```bash
go run main.go
```

服务将启动在 `http://localhost:8080`

### 4. 查看 API 文档

访问 Swagger 文档：
```
http://localhost:8080/swagger/index.html
```

## 主要依赖

- **Gin**: Web 框架
- **GORM**: ORM
- **PostgreSQL**: 数据库
- **JWT**: 身份认证
- **Viper**: 配置管理
- **Swagger**: API 文档

## 开发指南

### 添加新的 API 接口

1. 在 `controllers/` 创建控制器
2. 在 `router/router.go` 中添加路由
3. 添加 Swagger 注释
4. 运行 `swag init` 生成文档

### Swagger 注释示例

```go
// @Summary 用户登录
// @Description 用户登录接口
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} Response{data=LoginResponse}
// @Failure 400 {object} Response
// @Router /api/v1/auth/login [post]
func Login(c *gin.Context) {
    // 实现逻辑
}
```

## 环境变量

可通过环境变量覆盖配置：

```bash
export SERVER_PORT=8080
export DATABASE_HOST=localhost
export JWT_SECRET=your-secret-key
```

## API 测试

使用 curl 测试健康检查：

```bash
curl http://localhost:8080/health
```

## 下一步

- [ ] 实现用户认证相关接口
- [ ] 实现笔记 CRUD 接口
- [ ] 集成 AI 服务
- [ ] 添加单元测试
- [ ] 部署配置
