package bootstrap

import (
	"techmemo/backend/dao"
	"techmemo/backend/database"
	"techmemo/backend/service"
)

type App struct {
	UserService *service.UserService
}

func InitApp() *App {
	userDao := dao.NewUserDao(database.Q)
	userService := service.NewUserService(userDao)

	return &App{
		UserService: userService,
	}
}
