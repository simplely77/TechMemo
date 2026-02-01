package router

import (
	"techmemo/backend/bootstrap"
	"techmemo/backend/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 设置路由
func SetupRouter(app *bootstrap.App) *gin.Engine {
	r := gin.Default()

	// 中间件,配置跨域和自定义日志
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 认证路由（无需 JWT）
		auth := v1.Group("/auth")
		{
			auth.POST("/register", nil) // TODO: 实现注册
			auth.POST("/login", nil)    // TODO: 实现登录
		}

		// 需要认证的路由
		authorized := v1.Group("")
		authorized.Use(middleware.JWTAuth())
		{
			// 用户信息
			authorized.GET("/auth/profile", nil) // TODO: 获取用户信息

			// 分类管理
			categories := authorized.Group("/categories")
			{
				categories.GET("", nil)        // 获取分类列表
				categories.POST("", nil)       // 创建分类
				categories.PUT("/:id", nil)    // 更新分类
				categories.DELETE("/:id", nil) // 删除分类
			}

			// 标签管理
			tags := authorized.Group("/tags")
			{
				tags.GET("", nil)        // 获取标签列表
				tags.POST("", nil)       // 创建标签
				tags.DELETE("/:id", nil) // 删除标签
			}

			// 笔记管理
			notes := authorized.Group("/notes")
			{
				notes.GET("", nil)                                   // 获取笔记列表
				notes.POST("", nil)                                  // 创建笔记
				notes.GET("/:id", nil)                               // 获取笔记详情
				notes.PUT("/:id", nil)                               // 更新笔记
				notes.DELETE("/:id", nil)                            // 删除笔记
				notes.GET("/:id/versions", nil)                      // 获取版本历史
				notes.POST("/:id/versions/:version_id/restore", nil) // 恢复版本
				notes.GET("/:id/summary", nil)                       // 获取 AI 摘要
				notes.POST("/:id/extract-knowledge", nil)            // 提取知识点
			}

			// 知识点管理
			knowledge := authorized.Group("/knowledge-points")
			{
				knowledge.GET("", nil)        // 获取知识点列表
				knowledge.GET("/:id", nil)    // 获取知识点详情
				knowledge.PUT("/:id", nil)    // 更新知识点
				knowledge.DELETE("/:id", nil) // 删除知识点
			}

			// 知识点关系
			authorized.POST("/knowledge-relations", nil) // 创建知识点关系

			// AI 处理
			ai := authorized.Group("/ai")
			{
				ai.POST("/process/note/:id", nil) // 触发笔记 AI 处理
				ai.GET("/logs", nil)              // 获取 AI 处理日志
			}

			// 思维导图
			mindmap := authorized.Group("/mindmap")
			{
				mindmap.POST("/generate", nil) // 生成思维导图
				mindmap.GET("/global", nil)    // 获取全局知识图谱
			}

			// 搜索与问答
			authorized.POST("/search/semantic", nil) // 语义搜索
			authorized.GET("/search/keyword", nil)   // 关键词搜索
			authorized.POST("/qa/ask", nil)          // 智能问答

			// 复习推荐
			review := authorized.Group("/review")
			{
				review.GET("/recommendations", nil) // 获取复习推荐
				review.POST("/record", nil)         // 记录复习结果
			}

			// 统计分析
			stats := authorized.Group("/stats")
			{
				stats.GET("/overview", nil)   // 获取统计概览
				stats.GET("/categories", nil) // 分类统计
			}
		}
	}

	return r
}
