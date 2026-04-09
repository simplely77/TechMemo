package handler

import (
	"techmemo/backend/common/errors"
	"techmemo/backend/common/response"
	"techmemo/backend/handler/dto"
	"techmemo/backend/service"

	"github.com/gin-gonic/gin"
)

// @Summary 语义搜索
// @Description 基于向量相似度搜索笔记或知识点
// @Tags 搜索
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.SemanticSearchReq true "搜索参数"
// @Success 200 {object} response.Response{data=dto.SemanticSearchResp} "搜索成功"
// @Router /api/v1/search/semantic [post]
func HandlerSemanticSearch(searchService *service.SearchService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.SemanticSearchReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		// 从 JWT 中间件获取 user_id
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

		// 调用 service 执行搜索
		resp, err := searchService.SemanticSearch(c.Request.Context(), &req, userID)
		if err != nil {
			response.FailErr(c, err)
			return
		}

		response.Success(c, resp)
	}
}

// @Summary 获取搜索历史
// @Description 分页返回当前用户的搜索历史，按最近搜索时间倒序
// @Tags 搜索
// @Security BearerAuth
// @Produce json
// @Param page query int false "页码，默认 1"
// @Param page_size query int false "每页条数，默认 20，最大 100"
// @Success 200 {object} response.Response{data=dto.GetSearchHistoryResp} "成功"
// @Router /api/v1/search/history [get]
func HandlerGetSearchHistory(searchService *service.SearchService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.GetSearchHistoryReq
		if err := c.ShouldBindQuery(&req); err != nil {
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

		resp, err := searchService.GetSearchHistory(c.Request.Context(), &req, userID)
		if err != nil {
			response.FailErr(c, err)
			return
		}
		response.Success(c, resp)
	}
}
