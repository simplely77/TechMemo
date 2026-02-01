package handler

import (
	"techmemo/backend/common/errors"
	"techmemo/backend/common/response"
	"techmemo/backend/handler/dto"
	"techmemo/backend/service"

	"github.com/gin-gonic/gin"
)

func HandlerRegister(c *gin.Context, userService *service.UserService) {
	var req dto.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errors.InvalidParam)
		return
	}
	err := userService.Register(c.Request.Context(), req.Username, req.Password)

	if err != nil {
		response.Fail(c, err.(errors.ErrorCode))
		return
	}

	response.Success(c, nil)

}

func HandlerLogin(c *gin.Context) {

}

func HandlerProfile(c *gin.Context) {

}
