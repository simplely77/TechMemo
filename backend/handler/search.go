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
