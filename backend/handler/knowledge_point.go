package handler

import (
	"strconv"
	"techmemo/backend/common/errors"
	"techmemo/backend/common/response"
	"techmemo/backend/handler/dto"
	"techmemo/backend/service"

	"github.com/gin-gonic/gin"
)

// @Summary 获取知识点列表
// @Description 获取知识点列表，支持按笔记ID、关键词、重要程度筛选
// @Tags 知识点
// @Security BearerAuth
// @Param source_note_id query int64 false "来源笔记ID"
// @Param keyword query string false "关键词"
// @Param min_importance query number false "最低重要程度"
// @Param page query int64 false "页码（默认 1）"
// @Param page_size query int64 false "每页数量（默认 20）"
// @Success 200 {object} response.Response{data=dto.GetKnowledgePointsResp} "获取知识点列表成功"
// @Router /api/v1/knowledge-points [get]
func HandlerGetKnowledgePoints(knowledgePointService *service.KnowledgePointService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.GetKnowledgePointsReq
		if err := c.ShouldBindQuery(&req); err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		userIDAny, exist := c.Get("user_id")
		if !exist {
			response.Fail(c, errors.Unauthorized)
			return
		}

		userID := userIDAny.(int64)

		resp, err := knowledgePointService.GetKnowledgePoints(c.Request.Context(), &req, userID)
		if err != nil {
			response.Fail(c, errors.InternalErr)
			return
		}

		response.Success(c, resp)
	}
}

// @Summary 获取知识点详情
// @Description 获取知识点详情，包含关联知识点信息
// @Tags 知识点
// @Security BearerAuth
// @Param id path int64 true "知识点ID"
// @Success 200 {object} response.Response{data=dto.GetKnowledgePointResp} "获取知识点详情成功"
// @Router /api/v1/knowledge-points/{id} [get]
func HandlerGetKnowledgePoint(knowledgePointService *service.KnowledgePointService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		userIDAny, exist := c.Get("user_id")
		if !exist {
			response.Fail(c, errors.Unauthorized)
			return
		}

		userID := userIDAny.(int64)

		resp, err := knowledgePointService.GetKnowledgePoint(c.Request.Context(), id, userID)
		if err != nil {
			response.Fail(c, errors.InternalErr)
			return
		}

		if resp == nil {
			response.Fail(c, errors.NotFound)
			return
		}

		response.Success(c, resp)
	}
}

// @Summary 更新知识点
// @Description 更新知识点信息
// @Tags 知识点
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int64 true "知识点ID"
// @Param body body dto.UpdateKnowledgePointReq true "更新参数"
// @Success 200 {object} response.Response "更新知识点成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "无权限"
// @Failure 500 {object} response.Response "内部错误"
// @Router /api/v1/knowledge-points/{id} [put]
func HandlerUpdateKnowledgePoint(knowledgePointService *service.KnowledgePointService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		var req dto.UpdateKnowledgePointReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		userIDAny, exist := c.Get("user_id")
		if !exist {
			response.Fail(c, errors.Unauthorized)
			return
		}

		userID := userIDAny.(int64)

		err = knowledgePointService.UpdateKnowledgePoint(c.Request.Context(), id, &req, userID)
		if err != nil {
			response.Fail(c, errors.InternalErr)
			return
		}

		response.Success(c, nil)
	}
}

// @Summary 删除知识点
// @Description 删除知识点
// @Tags 知识点
// @Security BearerAuth
// @Param id path int64 true "知识点ID"
// @Success 200 {object} response.Response "删除知识点成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "无权限"
// @Failure 500 {object} response.Response "内部错误"
// @Router /api/v1/knowledge-points/{id} [delete]
func HandlerDeleteKnowledgePoint(knowledgePointService *service.KnowledgePointService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		userIDAny, exist := c.Get("user_id")
		if !exist {
			response.Fail(c, errors.Unauthorized)
			return
		}

		userID := userIDAny.(int64)

		err = knowledgePointService.DeleteKnowledgePoint(c.Request.Context(), id, userID)
		if err != nil {
			response.Fail(c, errors.InternalErr)
			return
		}

		response.Success(c, nil)
	}
}
