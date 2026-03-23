package handler

import (
	"techmemo/backend/common/errors"
	"techmemo/backend/common/response"
	"techmemo/backend/service"

	"github.com/gin-gonic/gin"
)

// @Summary 获取数据总览
// @Description 获取数据总览
// @Tags 数据
// @Security BearerAuth
// @Success 200 {object} response.Response{data=dto.GetOverviewStatsResp} "获取成功"
// @Router /api/v1/stats/overview [get]
func HandlerOverview(statsService *service.StatsService) gin.HandlerFunc {
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
		resp, err := statsService.GetOverview(c.Request.Context(), userID)
		if err != nil {
			response.Fail(c, errors.InternalErr)
			return
		}

		response.Success(c, resp)
	}
}

// @Summary 获取分类总览
// @Description 获取分类总览
// @Tags 数据
// @Security BearerAuth
// @Success 200 {object} response.Response{data=dto.GetCategoriesStatsResp} "获取成功"
// @Router /api/v1/stats/categories [get]
func HandlerCategories(statsService *service.StatsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDAny, exists := c.Get("user_id")
		if !exists {
			response.Fail(c, errors.Unauthorized)
			return
		}

		userID, ok := userIDAny.(int64)
		if !ok {
			response.Fail(c, errors.InternalErr)
			return
		}
		resp, err := statsService.GetCategories(c.Request.Context(), userID)
		if err != nil {
			response.Fail(c, errors.InternalErr)
			return
		}

		response.Success(c, resp)
	}
}
