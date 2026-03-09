package router

import (
	"techmemo/backend/bootstrap"
	"techmemo/backend/handler"
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
	// 将 docs 目录暴露为静态资源
	r.Static("/swagger", "./docs")

	// Swagger 文档路由
	r.GET("/swagger-ui/*filepath", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/swagger.json")))

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 认证路由（无需 JWT）
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handler.HandlerRegister(app.UserService))
			auth.POST("/login", handler.HandlerLogin(app.UserService))
		}

		// 需要认证的路由
		authorized := v1.Group("")
		authorized.Use(middleware.JWTAuth(app.UserService))
		{
			// 用户信息
			authorized.GET("/auth/profile", handler.HandlerProfile(app.UserService))

			// 分类管理
			categories := authorized.Group("/categories")
			{
				categories.GET("", handler.HandlerGetCategorys(app.CategoryService))          // 获取分类列表
				categories.POST("", handler.HandlerCreateCategory(app.CategoryService))       // 创建分类
				categories.PUT("/:id", handler.HandlerUpdateCategory(app.CategoryService))    // 更新分类
				categories.DELETE("/:id", handler.HandlerDeleteCategory(app.CategoryService)) // 删除分类
			}

			// 标签管理
			tags := authorized.Group("/tags")
			{
				tags.GET("", handler.HandlerGetTags(app.TagService))          // 获取标签列表
				tags.POST("", handler.HandlerCreateTag(app.TagService))       // 创建标签
				tags.PUT("/:id", handler.HandlerUpdateTag(app.TagService))    // 更新标签
				tags.DELETE("/:id", handler.HandlerDeleteTag(app.TagService)) // 删除标签
			}

			// 笔记管理
			notes := authorized.Group("/notes")
			{
				notes.GET("", handler.HandlerGetNotes(app.NoteService))       // 获取笔记列表
				notes.POST("", handler.HandlerCreateNote(app.NoteService))    // 创建笔记
				notes.GET("/:id", handler.HandlerGetNote(app.NoteService))    // 获取笔记详情
				notes.PUT("/:id", handler.HandlerUpdateNote(app.NoteService)) // 更新笔记
				notes.PUT("/:id/tags", handler.HandlerUpdateNoteTags(app.NoteService))
				notes.DELETE("/:id", handler.HandlerDeleteNote(app.NoteService))                             // 删除笔记
				notes.GET("/:id/versions", handler.HandlerGetNoteVersions(app.NoteService))                  // 获取版本历史
				notes.POST("/:id/versions/:version_id/restore", handler.HandlerRestoreNote(app.NoteService)) // 恢复版本                                                   // 提取知识点
			}

			// 知识点管理
			knowledge := authorized.Group("/knowledge-points")
			{
				knowledge.GET("", handler.HandlerGetKnowledgePoints(app.KnowledgePointService))          // 获取知识点列表
				knowledge.GET("/:id", handler.HandlerGetKnowledgePoint(app.KnowledgePointService))       // 获取知识点详情
				knowledge.PUT("/:id", handler.HandlerUpdateKnowledgePoint(app.KnowledgePointService))    // 更新知识点
				knowledge.DELETE("/:id", handler.HandlerDeleteKnowledgePoint(app.KnowledgePointService)) // 删除知识点
			}

			// AI 处理
			ai := authorized.Group("/ai")
			{
				ai.POST("/note/:id", handler.HandlerProcessNoteAI(app.NoteService, app.AIService)) // 触发笔记 AI 处理
				ai.GET("/note/:id/status", handler.HandlerGetNoteAIStatus(app.AIService))          // 获取 AI 处理日志
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
