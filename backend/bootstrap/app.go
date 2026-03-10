package bootstrap

import (
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
}

func InitApp() *App {
	userDao := dao.NewUserDao(database.Q)
	userService := service.NewUserService(userDao)
	categoryDao := dao.NewCategoryDao(database.Q)
	categoryService := service.NewCategoryService(categoryDao)
	tagDao := dao.NewTagDao(database.Q)
	tagService := service.NewTagService(tagDao)
	noteDao := dao.NewNoteDao(database.Q)
	noteService := service.NewNoteService(noteDao, categoryDao, tagDao, database.Q)
	aiDao := dao.NewAIDao(database.Q, database.DB)
	aiService := service.NewAIService(aiDao, noteDao)
	knowledgePointDao := dao.NewKnowledgePointDao(database.Q)
	knowledgePointService := service.NewKnowledgePointService(knowledgePointDao, noteDao)

	return &App{
		UserService:           userService,
		CategoryService:       categoryService,
		TagService:            tagService,
		NoteService:           noteService,
		AIService:             aiService,
		KnowledgePointService: knowledgePointService,
	}
}
