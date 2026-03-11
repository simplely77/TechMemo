package handler

import (
	"techmemo/backend/common/errors"
	"techmemo/backend/common/response"
	"techmemo/backend/handler/dto"
	"techmemo/backend/service"

	"github.com/gin-gonic/gin"
)

// @Summary 获取思维导图
// @Description 获取指定笔记的知识点树形结构（思维导图）
// @Tags 思维导图
// @Security BearerAuth
// @Param note_id query int64 true "笔记ID"
// @Success 200 {object} response.Response{data=dto.GetMindMapResp} "获取成功"
// @Router /api/v1/mindmap [get]
func HandlerGetMindMap(aiService *service.AIService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.GetMindMapReq
		if err := c.ShouldBindQuery(&req); err != nil {
			response.Fail(c, errors.InvalidParam)
			return
		}

		resp, err := aiService.GetMindMap(c.Request.Context(), req.NoteID)
		if err != nil {
			response.Fail(c, errors.InternalErr)
			return
		}

		response.Success(c, resp)
	}
}
