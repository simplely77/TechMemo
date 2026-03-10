package handler

import (
	"strconv"
	"techmemo/backend/common/errors"
	"techmemo/backend/common/response"
	"techmemo/backend/handler/dto"
	"techmemo/backend/service"

	"github.com/gin-gonic/gin"
)

// @Summary 获取分类
// @Description 获取用户全部分类
// @Tags 分类
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=dto.GetCategoriesResp} "获取分类成功"
// @Router /api/v1/categories [get]
func HandlerGetCategorys(categoryService *service.CategoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDAny, exist := c.Get("user_id")
		if !exist {
			response.Fail(c, errors.Unauthorized)
			return
		}

		userID, ok := userIDAny.(int64)
		if !ok {
			response.Fail(c, errors.InternalErr)
			return
		}

		resp, err := categoryService.GetCategorys(c.Request.Context(), userID)
		if err != nil {
			response.FailErr(c, err)
			return
		}

		response.Success(c, resp)
	}
}

// @Summary 创建分类
// @Description 创建分类
// @Tags 分类
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param data body dto.CreateCategoryReq true "创建分类参数"
// @Success 200 {object} response.Response{data=dto.CreateCategoryResp} "创建分类成功"
// @Router /api/v1/categories [post]
func HandlerCreateCategory(categoryService *service.CategoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.CreateCategoryReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		if req.Name == "" {
			response.Fail(c, errors.InvalidParam)
			return
		}

		userIDAny, exist := c.Get("user_id")
		if !exist {
			response.Fail(c, errors.Unauthorized)
			return
		}

		userID, ok := userIDAny.(int64)
		if !ok {
			response.Fail(c, errors.InternalErr)
			return
		}

		resp, err := categoryService.CreateCategory(c.Request.Context(), userID, req.Name)
		if err != nil {
			response.FailErr(c, err)
			return
		}

		response.Success(c, resp)
	}
}

// @Summary 更新分类
// @Description 更新分类
// @Tags 分类
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "分类ID"
// @Param data body dto.UpdateCategoryReq true "更新分类参数"
// @Success 200 {object} response.Response "更新分类成功"
// @Router /api/v1/categories/{id} [put]
func HandlerUpdateCategory(categoryService *service.CategoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDAny, exist := c.Get("user_id")
		if !exist {
			response.Fail(c, errors.Unauthorized)
			return
		}

		userID, ok := userIDAny.(int64)
		if !ok {
			response.Fail(c, errors.InternalErr)
			return
		}
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		var req dto.UpdateCategoryReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		err = categoryService.UpdateCategory(c.Request.Context(), userID, id, req.Name)
		if err != nil {
			response.FailErr(c, err)
			return
		}
		response.Success(c, nil)
	}
}

// @Summary 删除分类
// @Description 删除分类
// @Tags 分类
// @Security BearerAuth
// @Param id path int true "分类ID"
// @Success 200 {object} response.Response "更新分类成功"
// @Router /api/v1/categories/{id} [delete]
func HandlerDeleteCategory(categoryService *service.CategoryService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		err = categoryService.DeleteCategory(c.Request.Context(), id)
		if err != nil {
			response.FailErr(c, err)
			return
		}
		response.Success(c, nil)
	}
}
