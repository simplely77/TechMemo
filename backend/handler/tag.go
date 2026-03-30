package handler

import (
	"strconv"
	"techmemo/backend/common/errors"
	"techmemo/backend/common/response"
	"techmemo/backend/handler/dto"
	"techmemo/backend/service"

	"github.com/gin-gonic/gin"
)

// @Summary 获取标签
// @Description 获取用户全部标签
// @Tags 标签
// @Security BearerAuth
// @Produce json
// @Param keyword query string false "标签名关键词搜索"
// @Param page query int false "页码（默认 1）"
// @Param page_size query int false "每页数量（默认 20）"
// @Success 200 {object} response.Response{data=dto.GetTagsResp} "获取标签成功"
// @Router /api/v1/tags [get]
func HandlerGetTags(tagService *service.TagService) gin.HandlerFunc {
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
		var req dto.GetTagsReq
		if err := c.ShouldBindQuery(&req); err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		if req.Page <= 0 {
			req.Page = 1
		}
		if req.PageSize <= 0 {
			req.PageSize = 20
		}
		if req.PageSize > 100 { // 限制最大分页
			req.PageSize = 100
		}
		resp, err := tagService.GetTags(c.Request.Context(), userID, req.Keyword, req.Page, req.PageSize)
		if err != nil {
			response.FailErr(c, err)
			return
		}

		response.Success(c, resp)
	}
}

// @Summary 创建标签
// @Description 创建标签
// @Tags 标签
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param data body dto.CreateTagReq true "创建标签参数"
// @Success 200 {object} response.Response{data=dto.Tag} "创建标签成功"
// @Router /api/v1/tags [post]
func HandlerCreateTag(tagService *service.TagService) gin.HandlerFunc {
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
		var req dto.CreateTagReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		tag, err := tagService.CreateTag(c.Request.Context(), userID, req.Name)
		if err != nil {
			response.FailErr(c, err)
			return
		}

		response.Success(c, tag)
	}
}

// @Summary 更新标签
// @Description 更新标签
// @Tags 标签
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "标签ID"
// @Param data body dto.UpdateTagReq true "更新标签参数"
// @Success 200 {object} response.Response{data=dto.Tag} "更新标签成功"
// @Router /api/v1/tags/{id} [put]
func HandlerUpdateTag(tagService *service.TagService) gin.HandlerFunc {
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

		var req dto.UpdateTagReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		resp, err := tagService.UpdateTag(c.Request.Context(), userID, id, req.Name)
		if err != nil {
			response.FailErr(c, err)
			return
		}
		response.Success(c, resp)
	}
}

// @Summary 删除标签
// @Description 删除标签
// @Tags 标签
// @Security BearerAuth
// @Produce json
// @Param id path int true "标签ID"
// @Success 200 {object} response.Response "删除标签成功"
// @Router /api/v1/tags/{id} [delete]
func HandlerDeleteTag(tagService *service.TagService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		err = tagService.DeleteTag(c.Request.Context(), id)
		if err != nil {
			response.FailErr(c, err)
			return
		}
		response.Success(c, nil)
	}
}
