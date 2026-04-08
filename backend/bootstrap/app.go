package bootstrap

import (
	aiclient "techmemo/backend/ai/client"
	"techmemo/backend/config"
	"techmemo/backend/dao"
	"techmemo/backend/database"
	"techmemo/backend/service"
)

type App struct {
	UserService           *service.UserService
	CategoryService       *service.CategoryService
	TagService            *service.TagService
	NoteService           *service.NoteService
	AIService             *service.AIService
	KnowledgePointService *service.KnowledgePointService
	SearchService         *service.SearchService
	StatsService          *service.StatsService
	ChatService           *service.ChatService
}

func InitApp() *App {
	userDao := dao.NewUserDao(database.Q)
	userService := service.NewUserService(userDao)
	tagDao := dao.NewTagDao(database.Q)
	tagService := service.NewTagService(tagDao)
	noteDao := dao.NewNoteDao(database.Q)
	chatDao := dao.NewChatDao(database.Q)
	aiDao := dao.NewAIDao(database.Q, database.DB)
	categoryDao := dao.NewCategoryDao(database.Q)
	categoryService := service.NewCategoryService(categoryDao, noteDao)

	knowledgePointDao := dao.NewKnowledgePointDao(database.Q)
	noteService := service.NewNoteService(noteDao, categoryDao, tagDao, knowledgePointDao, database.Q)
	knowledgePointService := service.NewKnowledgePointService(knowledgePointDao, noteDao)

	// 初始化 SearchService
	searchDao := dao.NewSearchDao(database.Q, database.DB)
	aiClient := aiclient.NewOpenAIClientFromConfig(config.AppConfig)
	searchService := service.NewSearchService(
		searchDao,
		noteDao,
		knowledgePointDao,
		categoryDao,
		aiClient,
	)

	aiService := service.NewAIService(
		aiDao,
		noteDao,
		aiClient,
	)

	statsService := service.NewStatsServcie(
		noteDao,
		categoryDao,
		knowledgePointDao,
		tagDao,
		aiDao,
	)

	// 初始化 ChatService
	chatService := service.NewChatService(
		chatDao,
		searchDao,
		noteDao,
		knowledgePointDao,
		aiClient,
	)

	return &App{
		UserService:           userService,
		CategoryService:       categoryService,
		TagService:            tagService,
		NoteService:           noteService,
		AIService:             aiService,
		KnowledgePointService: knowledgePointService,
		SearchService:         searchService,
		StatsService:          statsService,
		ChatService:           chatService,
	}
}
