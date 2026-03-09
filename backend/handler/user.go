package handler

import (
	"techmemo/backend/common/errors"
	"techmemo/backend/common/response"
	"techmemo/backend/handler/dto"
	"techmemo/backend/service"

	"github.com/gin-gonic/gin"
)

// @Summary 用户注册
// @Description 创建新用户账号
// @Tags 授权
// @Accept json
// @Produce json
// @Param data body dto.RegisterReq true "注册参数"
// @Success 200 {object} response.Response{data=dto.RegisterResp} "注册成功"
// @Router /api/v1/auth/register [post]
func HandlerRegister(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.RegisterReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, errors.InvalidParam) // 参数错误
			return
		}

		// 对用户名密码做简单校验
		if req.Username == "" || req.Password == "" {
			response.Fail(c, errors.InvalidParam)
			return
		}
		// 判断用户是否已存在
		resp, err := userService.Register(c.Request.Context(), req)
		if err != nil {
			response.Fail(c, err.(errors.ErrorCode)) // 返回具体的错误
			return
		}

		// 返回成功响应
		response.Success(c, resp) // 成功时返回 RegisterResp 数据
	}
}

// @Summary 用户登录
// @Description 用户名密码登录
// @Tags 授权
// @Accept json
// @Produce json
// @Param data body dto.LoginReq true "登录参数"
// @Success 200 {object} response.Response{data=dto.LoginResp} "登录成功"
// @Router /api/v1/auth/login [post]
func HandlerLogin(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.LoginReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		// 对用户名密码做简单校验
		if req.Username == "" || req.Password == "" {
			response.Fail(c, errors.InvalidParam)
			return
		}

		// 调用 Service 层进行用户登录校验
		resp, err := userService.Login(c.Request.Context(), req)
		if err != nil {
			response.Fail(c, err.(errors.ErrorCode))
			return
		}

		// 返回登录成功的响应
		response.Success(c, resp)
	}
}

// @Summary 用户信息
// @Description 获取用户信息
// @Tags 授权
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=dto.ProfileResp} "获取用户信息成功"
// @Router /api/v1/auth/profile [get]
func HandlerProfile(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDAny, exist := c.Get("user_id")
		if !exist {
			response.Fail(c, errors.Unauthorized)
			return
		}

		userID, ok := userIDAny.(int64)
		if !ok {
			response.Fail(c, errors.InternalErr)
		}

		resp, err := userService.GetProfile(c.Request.Context(), userID)
		if err != nil {
			response.Fail(c, err.(errors.ErrorCode))
			return
		}

		response.Success(c, resp)
	}
}
